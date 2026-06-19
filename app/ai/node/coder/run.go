package coder

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MoScenix/ai-code/app/ai/agent"
	"github.com/MoScenix/ai-code/app/ai/utils"
	"github.com/MoScenix/ai-code/common/aievent"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/kitex/pkg/klog"
)

const agentName = "Coder"

const terminalTaskTTL = 10 * time.Second

func Run(ctx context.Context, input map[string]any) (map[string]any, error) {
	if utils.IsCancelled(ctx) {
		return map[string]any{}, nil
	}

	store, ok := utils.ProjectStoreFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("coder node requires project store")
	}

	projectID, _ := utils.ProjectIDFromContext(ctx)
	streamStore, _ := utils.StreamStoreFromContext(ctx)
	stateStore, _ := utils.StateStoreFromContext(ctx)
	initialMessages := historyMessages(ctx, input)
	if len(initialMessages) == 0 {
		klog.CtxWarnf(ctx, "skip coder task: project_id=%s reason=empty_initial_messages", projectID)
		return map[string]any{}, nil
	}
	klog.CtxInfof(ctx, "coder task started: project_id=%s", projectID)

	coderAgent, err := agent.NewCoder(ctx, store)
	if err != nil {
		return nil, err
	}

	var lastEventID string
	assistantOutput := &utils.StringBuffer{}
	loopCtx, cancelLoop := context.WithCancel(ctx)
	defer cancelLoop()

	loop := adk.NewTurnLoop[[]*schema.Message, *schema.Message](adk.TurnLoopConfig[[]*schema.Message, *schema.Message]{
		GenInput: genInput,
		PrepareAgent: func(context.Context, *adk.TurnLoop[[]*schema.Message, *schema.Message], [][]*schema.Message) (adk.TypedAgent[*schema.Message], error) {
			return coderAgent, nil
		},
		OnAgentEvents: func(ctx context.Context, _ *adk.TurnContext[[]*schema.Message, *schema.Message], events *adk.AsyncIterator[*adk.TypedAgentEvent[*schema.Message]]) error {
			return publishAgentEvents(ctx, streamStore, projectID, events, &lastEventID, assistantOutput)
		},
	})

	lastEventID, _ = publishTaskEvent(ctx, streamStore, aievent.TaskEvent{
		ProjectID: projectID,
		Type:      aievent.EventAgentStart,
		Agent:     agentName,
		CreatedAt: time.Now().UnixMilli(),
	})
	_ = setProjectState(ctx, stateStore, projectID, aievent.ProjectState{
		Status:      aievent.ProjectStatusRunning,
		Agent:       agentName,
		LastEventID: projectLastEventID(ctx, stateStore, projectID),
		UpdatedAt:   time.Now().UnixMilli(),
	})
	loop.Push(initialMessages)
	loop.Stop(adk.UntilIdleFor(3 * time.Second))

	controlCtx, cancelControl := context.WithCancel(ctx)
	controlDone := make(chan struct{})
	go func() {
		defer close(controlDone)
		watchStream(controlCtx, stateStore, streamStore, projectID, loop, &lastEventID, assistantOutput)
	}()

	loop.Run(loopCtx)
	state := loop.Wait()
	cancelLoop()
	cancelControl()
	<-controlDone
	output := assistantOutput.String()

	if state != nil && state.ExitReason != nil {
		if state.StopCause != "" {
			klog.CtxWarnf(ctx, "coder task cancelled: project_id=%s cause=%s", projectID, state.StopCause)
			publishTerminalEvent(ctx, streamStore, stateStore, projectID, aievent.EventCancelled, aievent.ProjectStatusCancelled, state.StopCause)
			return map[string]any{}, nil
		}
		klog.CtxErrorf(ctx, "coder task failed: project_id=%s err=%v", projectID, state.ExitReason)
		publishTerminalEvent(ctx, streamStore, stateStore, projectID, aievent.EventError, aievent.ProjectStatusError, state.ExitReason.Error())
		return nil, state.ExitReason
	}

	if err := store.Commit(ctx); err != nil {
		klog.CtxErrorf(ctx, "commit coder changes failed: project_id=%s err=%v", projectID, err)
		publishTerminalEvent(ctx, streamStore, stateStore, projectID, aievent.EventError, aievent.ProjectStatusError, err.Error())
		return nil, err
	}

	if err := utils.AddProjectAssistantMessage(ctx, projectID, output); err != nil {
		klog.CtxErrorf(ctx, "persist coder assistant message failed: project_id=%s err=%v", projectID, err)
		publishTerminalEvent(ctx, streamStore, stateStore, projectID, aievent.EventError, aievent.ProjectStatusError, err.Error())
		return nil, err
	}
	if strings.TrimSpace(output) != "" {
		_ = updateProjectLastEventID(ctx, stateStore, projectID, lastEventID)
	}

	publishTerminalEvent(ctx, streamStore, stateStore, projectID, aievent.EventDone, aievent.ProjectStatusDone, "")
	klog.CtxInfof(ctx, "coder task completed: project_id=%s", projectID)
	return map[string]any{}, nil
}

func genInput(_ context.Context, _ *adk.TurnLoop[[]*schema.Message, *schema.Message], items [][]*schema.Message) (*adk.GenInputResult[[]*schema.Message, *schema.Message], error) {
	messages := make([]*schema.Message, 0)
	for _, item := range items {
		messages = append(messages, item...)
	}
	if len(messages) == 0 {
		return nil, fmt.Errorf("coder node received empty input")
	}
	return &adk.GenInputResult[[]*schema.Message, *schema.Message]{
		Input: &adk.TypedAgentInput[*schema.Message]{
			Messages: messages,
		},
		Consumed: items,
	}, nil
}

func historyMessages(ctx context.Context, _ map[string]any) []*schema.Message {
	history, _ := utils.HistoryMessagesFromContext(ctx)
	messages := make([]*schema.Message, 0, len(history)+1)
	for _, msg := range history {
		if msg != nil {
			messages = append(messages, msg)
		}
	}
	if buffer, ok := utils.StringBufferFromContext(ctx); ok {
		if extra := strings.TrimSpace(buffer.String()); extra != "" {
			messages = append(messages, schema.AssistantMessage(extra, nil))
		}
	}
	return messages
}
