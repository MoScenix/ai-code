package designer

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MoScenix/ai-code/app/ai/agent"
	"github.com/MoScenix/ai-code/app/ai/utils"
	"github.com/MoScenix/ai-code/common/aievent"
	"github.com/MoScenix/ai-code/common/redisstate"
	"github.com/MoScenix/ai-code/common/redisstream"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

const agentName = "Designer"

var ErrInterrupted = errors.New("designer interrupted waiting for user answer")
var ErrNoInterruptedDesigner = errors.New("designer has no interrupted checkpoint")

type InterruptedState struct {
	CheckpointID      string
	Checkpoint        []byte
	PendingInterrupts []aievent.PendingInterrupt
	Buffer            string
	LastEventID       string
	ControlCursor     string
}

type designerSession struct {
	ctx          context.Context
	loopCtx      context.Context
	cancel       context.CancelFunc
	agent        *adk.ChatModelAgent
	projectID    string
	streamStore  redisstream.Store
	stateStore   *redisstate.Store
	buffer       *utils.StringBuffer
	checkpointID string
	checkpoints  *memoryCheckpointStore
	lastEventID  string
	answers      chan answerEvent
	design       strings.Builder
}

func init() {
	schema.RegisterName[InterruptedState]("ai_designer_interrupted_state_v1")
}

func Run(ctx context.Context) (map[string]any, error) {
	if utils.IsCancelled(ctx) {
		return map[string]any{}, nil
	}

	if wasInterrupted, hasState, interrupted := compose.GetInterruptState[InterruptedState](ctx); wasInterrupted {
		if !hasState {
			return nil, ErrNoInterruptedDesigner
		}
		isResume, hasData, answer := compose.GetResumeContext[agent.DesignerAnswer](ctx)
		if !isResume || !hasData {
			return nil, compose.StatefulInterrupt(ctx, graphInterruptInfo(interrupted), interrupted)
		}
		return runResumed(ctx, interrupted, answer)
	}

	initialMessages := historyMessages(ctx)
	if len(initialMessages) == 0 {
		return map[string]any{}, nil
	}

	projectID, _ := utils.ProjectIDFromContext(ctx)
	checkpointID := designerCheckpointID(projectID)
	session, err := newDesignerSession(ctx, checkpointID, newMemoryCheckpointStore())
	if err != nil {
		return nil, err
	}
	defer session.close()

	session.lastEventID, _ = publishTaskEvent(ctx, session.streamStore, aievent.TaskEvent{
		ProjectID: session.projectID,
		Type:      aievent.EventAgentStart,
		Agent:     agentName,
		CreatedAt: time.Now().UnixMilli(),
	})
	_ = setProjectState(ctx, session.stateStore, session.projectID, aievent.ProjectState{
		Status:      "running",
		Agent:       agentName,
		LastEventID: stateLastEventID(ctx, session.stateStore, session.projectID),
		UpdatedAt:   time.Now().UnixMilli(),
	})
	return session.run(initialMessages, nil)
}

func runResumed(ctx context.Context, interrupted InterruptedState, answer agent.DesignerAnswer) (map[string]any, error) {
	if utils.IsCancelled(ctx) {
		return map[string]any{}, nil
	}

	buffer, _ := utils.StringBufferFromContext(ctx)
	if buffer != nil {
		buffer.SetString(interrupted.Buffer)
	}
	if interrupted.ControlCursor != "" {
		utils.SetControlCursor(ctx, interrupted.ControlCursor)
	}

	checkpointID := interrupted.CheckpointID
	checkpoints := newMemoryCheckpointStore()
	if checkpointID == "" || len(interrupted.Checkpoint) == 0 {
		return nil, ErrNoInterruptedDesigner
	}
	if err := checkpoints.Set(ctx, checkpointID, interrupted.Checkpoint); err != nil {
		return nil, err
	}

	session, err := newDesignerSession(ctx, checkpointID, checkpoints)
	if err != nil {
		return nil, err
	}
	defer session.close()

	session.lastEventID, _ = publishTaskEvent(ctx, session.streamStore, aievent.TaskEvent{
		ProjectID: session.projectID,
		Type:      aievent.EventAccepted,
		Agent:     agentName,
		Content:   "designer resume accepted",
		CreatedAt: time.Now().UnixMilli(),
	})
	_ = setProjectState(ctx, session.stateStore, session.projectID, aievent.ProjectState{
		Status:       "running",
		Agent:        agentName,
		LastEventID:  stateLastEventID(ctx, session.stateStore, session.projectID),
		CheckpointID: checkpointID,
		Buffer:       bufferValue(buffer),
		IsCancelled:  utils.IsCancelled(ctx),
		UpdatedAt:    time.Now().UnixMilli(),
	})
	return session.run(nil, &adk.ResumeParams{
		Targets: resumeTargets(interrupted.PendingInterrupts, answer),
	})
}

func newDesignerSession(ctx context.Context, checkpointID string, checkpoints *memoryCheckpointStore) (*designerSession, error) {
	designerAgent, err := agent.NewDesigner(ctx)
	if err != nil {
		return nil, err
	}
	loopCtx, cancel := context.WithCancel(ctx)
	return &designerSession{
		ctx:          ctx,
		loopCtx:      loopCtx,
		cancel:       cancel,
		projectID:    projectID(ctx),
		streamStore:  streamStore(ctx),
		stateStore:   stateStore(ctx),
		buffer:       stringBuffer(ctx),
		checkpointID: checkpointID,
		checkpoints:  checkpoints,
		answers:      make(chan answerEvent, 8),
		agent:        designerAgent,
	}, nil
}

func (s *designerSession) close() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *designerSession) run(initialMessages []*schema.Message, resumeParams *adk.ResumeParams) (map[string]any, error) {
	for {
		interrupt, cleanup, err := s.runTurn(initialMessages, resumeParams)
		initialMessages = nil
		resumeParams = nil
		if err != nil {
			cleanup()
			return nil, err
		}
		if interrupt == nil {
			cleanup()
			break
		}

		answer, ok, err := waitAnswer(s.ctx, s.answers, interrupt.ID)
		cleanup()
		if err != nil {
			return nil, err
		}
		if !ok {
			interrupted, err := buildInterruptedState(s.ctx, s.checkpoints, s.checkpointID, s.lastEventID, interrupt, s.buffer)
			if err != nil {
				return nil, err
			}
			return nil, compose.StatefulInterrupt(s.ctx, graphInterruptInfo(interrupted), interrupted)
		}

		_ = setProjectState(s.ctx, s.stateStore, s.projectID, aievent.ProjectState{
			Status:       aievent.ProjectStatusRunning,
			Agent:        agentName,
			LastEventID:  stateLastEventID(s.ctx, s.stateStore, s.projectID),
			CheckpointID: s.checkpointID,
			Buffer:       bufferValue(s.buffer),
			IsCancelled:  utils.IsCancelled(s.ctx),
			UpdatedAt:    time.Now().UnixMilli(),
		})
		resumeParams = &adk.ResumeParams{
			Targets: map[string]any{
				interrupt.ID: answer,
			},
		}
	}

	_ = setProjectState(s.ctx, s.stateStore, s.projectID, aievent.ProjectState{
		Status:      "running",
		Agent:       agentName,
		LastEventID: stateLastEventID(s.ctx, s.stateStore, s.projectID),
		Buffer:      bufferValue(s.buffer),
		UpdatedAt:   time.Now().UnixMilli(),
	})
	if s.stateStore != nil && s.projectID != "" {
		_ = s.stateStore.Del(s.ctx, aievent.CheckpointKey(s.projectID))
	}

	if s.buffer != nil {
		s.buffer.WriteString(s.design.String())
	}
	return map[string]any{}, nil
}

func (s *designerSession) runTurn(initialMessages []*schema.Message, resumeParams *adk.ResumeParams) (*interruptEvent, func(), error) {
	var interrupt *interruptEvent
	loop := adk.NewTurnLoop[[]*schema.Message, *schema.Message](adk.TurnLoopConfig[[]*schema.Message, *schema.Message]{
		Store:        s.checkpoints,
		CheckpointID: s.checkpointID,
		GenInput:     genDesignerInput,
		GenResume: func(_ context.Context, _ *adk.TurnLoop[[]*schema.Message, *schema.Message], interruptedItems, unhandledItems, newItems [][]*schema.Message) (*adk.GenResumeResult[[]*schema.Message, *schema.Message], error) {
			items := append(append(interruptedItems, unhandledItems...), newItems...)
			return &adk.GenResumeResult[[]*schema.Message, *schema.Message]{
				ResumeParams: resumeParams,
				Consumed:     items,
			}, nil
		},
		PrepareAgent: func(context.Context, *adk.TurnLoop[[]*schema.Message, *schema.Message], [][]*schema.Message) (adk.TypedAgent[*schema.Message], error) {
			return s.agent, nil
		},
		OnAgentEvents: func(ctx context.Context, _ *adk.TurnContext[[]*schema.Message, *schema.Message], events *adk.AsyncIterator[*adk.TypedAgentEvent[*schema.Message]]) error {
			nextInterrupt, content, err := publishAgentEvents(ctx, s.streamStore, s.projectID, events, &s.lastEventID)
			if content != "" {
				s.design.WriteString(content)
			}
			if nextInterrupt != nil {
				interrupt = nextInterrupt
			}
			return err
		},
	})

	watchCtx, cancelWatch := context.WithCancel(s.ctx)
	watchDone := make(chan struct{})
	go func() {
		defer close(watchDone)
		watchPushes(watchCtx, s.stateStore, s.streamStore, s.projectID, s.buffer, s.answers, loop)
	}()

	cleanup := func() {
		cancelWatch()
		<-watchDone
	}

	if initialMessages != nil {
		loop.Push(initialMessages)
	}
	loop.Stop(adk.UntilIdleFor(3 * time.Second))

	loop.Run(s.loopCtx)
	state := loop.Wait()
	if state != nil && state.ExitReason != nil && interrupt == nil {
		if state.StopCause != "" {
			utils.CancelRuntime(s.ctx)
			return nil, cleanup, nil
		}
		return nil, cleanup, state.ExitReason
	}
	return interrupt, cleanup, nil
}

func genDesignerInput(_ context.Context, _ *adk.TurnLoop[[]*schema.Message, *schema.Message], items [][]*schema.Message) (*adk.GenInputResult[[]*schema.Message, *schema.Message], error) {
	messages := make([]*schema.Message, 0)
	for _, item := range items {
		messages = append(messages, item...)
	}
	if len(messages) == 0 {
		return nil, fmt.Errorf("designer node received empty input")
	}
	return &adk.GenInputResult[[]*schema.Message, *schema.Message]{
		Input: &adk.TypedAgentInput[*schema.Message]{
			Messages: messages,
		},
		Consumed: items,
	}, nil
}

func projectID(ctx context.Context) string {
	projectID, _ := utils.ProjectIDFromContext(ctx)
	return projectID
}

func streamStore(ctx context.Context) redisstream.Store {
	store, _ := utils.StreamStoreFromContext(ctx)
	return store
}

func stateStore(ctx context.Context) *redisstate.Store {
	store, _ := utils.StateStoreFromContext(ctx)
	return store
}

func stringBuffer(ctx context.Context) *utils.StringBuffer {
	buffer, _ := utils.StringBufferFromContext(ctx)
	return buffer
}

func buildInterruptedState(ctx context.Context, checkpoints *memoryCheckpointStore, checkpointID string, lastEventID string, interrupt *interruptEvent, buffer *utils.StringBuffer) (InterruptedState, error) {
	data, existed, err := checkpoints.Get(ctx, checkpointID)
	if err != nil {
		return InterruptedState{}, err
	}
	if !existed {
		return InterruptedState{}, fmt.Errorf("designer checkpoint %q not found", checkpointID)
	}
	return InterruptedState{
		CheckpointID:  checkpointID,
		Checkpoint:    data,
		LastEventID:   lastEventID,
		ControlCursor: utils.ControlCursor(ctx),
		PendingInterrupts: []aievent.PendingInterrupt{
			{
				ID:      interrupt.ID,
				Agent:   agentName,
				Content: interrupt.Content,
				Payload: interrupt.Payload,
			},
		},
		Buffer: bufferValue(buffer),
	}, nil
}

func designerCheckpointID(projectID string) string {
	if projectID == "" {
		return "designer"
	}
	return "project:" + projectID + ":designer"
}

func historyMessages(ctx context.Context) []*schema.Message {
	history, _ := utils.HistoryMessagesFromContext(ctx)
	messages := make([]*schema.Message, 0, len(history)+1)
	for _, msg := range history {
		if msg != nil {
			messages = append(messages, msg)
		}
	}
	if buffer, ok := utils.StringBufferFromContext(ctx); ok {
		if extra := strings.TrimSpace(buffer.String()); extra != "" {
			messages = append(messages, schema.SystemMessage("Pending designer input:\n"+extra))
		}
	}
	return messages
}

func bufferValue(buffer *utils.StringBuffer) string {
	if buffer == nil {
		return ""
	}
	return buffer.String()
}

func resumeTargets(interrupts []aievent.PendingInterrupt, answer agent.DesignerAnswer) map[string]any {
	targets := make(map[string]any, len(interrupts))
	for _, interrupt := range interrupts {
		if interrupt.ID != "" {
			targets[interrupt.ID] = answer
		}
	}
	return targets
}

func graphInterruptInfo(interrupted InterruptedState) any {
	if len(interrupted.PendingInterrupts) == 0 {
		return map[string]any{
			"agent": agentName,
		}
	}
	pending := interrupted.PendingInterrupts[0]
	return map[string]any{
		"agent":                       agentName,
		"content":                     pending.Content,
		"payload":                     pending.Payload,
		aievent.PayloadADKInterruptID: pending.ID,
		"adk_checkpoint_id":           interrupted.CheckpointID,
		aievent.PayloadDesignerLastID: interrupted.LastEventID,
		aievent.PayloadControlCursor:  interrupted.ControlCursor,
		"designer_has_state":          len(interrupted.Checkpoint) > 0,
	}
}
