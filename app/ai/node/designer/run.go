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
}

type designerSession struct {
	ctx          context.Context
	loopCtx      context.Context
	cancel       context.CancelFunc
	pushDone     chan struct{}
	projectID    string
	streamStore  redisstream.Store
	stateStore   *redisstate.Store
	buffer       *utils.StringBuffer
	checkpointID string
	checkpoints  *memoryCheckpointStore
	runner       *adk.Runner
	lastEventID  string
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
		LastEventID: session.lastEventID,
		UpdatedAt:   time.Now().UnixMilli(),
	})
	session.watchPushes()

	events := session.runner.Run(session.loopCtx, initialMessages, adk.WithCheckPointID(checkpointID))
	return session.consume(events)
}

func runResumed(ctx context.Context, interrupted InterruptedState, answer agent.DesignerAnswer) (map[string]any, error) {
	if utils.IsCancelled(ctx) {
		return map[string]any{}, nil
	}

	buffer, _ := utils.StringBufferFromContext(ctx)
	if buffer != nil {
		buffer.SetString(interrupted.Buffer)
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
		LastEventID:  session.lastEventID,
		CheckpointID: checkpointID,
		Buffer:       bufferValue(buffer),
		IsCancelled:  utils.IsCancelled(ctx),
		UpdatedAt:    time.Now().UnixMilli(),
	})
	session.watchPushes()

	events, err := session.runner.ResumeWithParams(session.loopCtx, checkpointID, &adk.ResumeParams{
		Targets: resumeTargets(interrupted.PendingInterrupts, answer),
	})
	if err != nil {
		return nil, err
	}
	return session.consume(events)
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
		runner: adk.NewRunner(ctx, adk.RunnerConfig{
			Agent:           designerAgent,
			EnableStreaming: true,
			CheckPointStore: checkpoints,
		}),
	}, nil
}

func (s *designerSession) watchPushes() {
	s.pushDone = make(chan struct{})
	go func() {
		defer close(s.pushDone)
		watchPushes(s.loopCtx, s.streamStore, s.projectID, &s.lastEventID, s.buffer)
	}()
}

func (s *designerSession) close() {
	if s.cancel != nil {
		s.cancel()
	}
	if s.pushDone != nil {
		<-s.pushDone
	}
}

func (s *designerSession) consume(events *adk.AsyncIterator[*adk.TypedAgentEvent[*schema.Message]]) (map[string]any, error) {
	var design string
	for {
		interrupt, content, err := publishAgentEvents(s.ctx, s.streamStore, s.projectID, events, &s.lastEventID)
		design += content
		if err != nil {
			return nil, err
		}
		if interrupt == nil {
			break
		}

		nextAnswer, ok, err := waitAnswer(s.ctx, s.streamStore, s.projectID, interrupt.EventID, interrupt.ID)
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

		nextEvents, err := s.runner.ResumeWithParams(s.loopCtx, s.checkpointID, &adk.ResumeParams{
			Targets: map[string]any{
				interrupt.ID: nextAnswer,
			},
		})
		if err != nil {
			return nil, err
		}
		events = nextEvents
	}

	s.lastEventID, _ = publishTaskEvent(s.ctx, s.streamStore, aievent.TaskEvent{
		ProjectID: s.projectID,
		Type:      aievent.EventDone,
		Agent:     agentName,
		Content:   "designer done",
		CreatedAt: time.Now().UnixMilli(),
	})
	_ = setProjectState(s.ctx, s.stateStore, s.projectID, aievent.ProjectState{
		Status:      "running",
		Agent:       agentName,
		LastEventID: s.lastEventID,
		Buffer:      bufferValue(s.buffer),
		UpdatedAt:   time.Now().UnixMilli(),
	})
	if s.stateStore != nil && s.projectID != "" {
		_ = s.stateStore.Del(s.ctx, aievent.CheckpointKey(s.projectID))
	}

	if s.buffer != nil {
		s.buffer.WriteString(design)
	}
	return map[string]any{}, nil
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
		CheckpointID: checkpointID,
		Checkpoint:   data,
		LastEventID:  lastEventID,
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
		"agent":              agentName,
		"content":            pending.Content,
		"payload":            pending.Payload,
		"adk_interrupt_id":   pending.ID,
		"adk_checkpoint_id":  interrupted.CheckpointID,
		"designer_last_id":   interrupted.LastEventID,
		"designer_has_state": len(interrupted.Checkpoint) > 0,
	}
}
