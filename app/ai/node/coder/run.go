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
)

const agentName = "Coder"

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
		return map[string]any{}, nil
	}

	coderAgent, err := agent.NewCoder(ctx, store)
	if err != nil {
		return nil, err
	}

	var lastEventID string
	loopCtx, cancelLoop := context.WithCancel(ctx)
	defer cancelLoop()

	loop := adk.NewTurnLoop[[]*schema.Message, *schema.Message](adk.TurnLoopConfig[[]*schema.Message, *schema.Message]{
		GenInput: genInput,
		PrepareAgent: func(context.Context, *adk.TurnLoop[[]*schema.Message, *schema.Message], [][]*schema.Message) (adk.TypedAgent[*schema.Message], error) {
			return coderAgent, nil
		},
		OnAgentEvents: func(ctx context.Context, _ *adk.TurnContext[[]*schema.Message, *schema.Message], events *adk.AsyncIterator[*adk.TypedAgentEvent[*schema.Message]]) error {
			return publishAgentEvents(ctx, streamStore, projectID, events, &lastEventID)
		},
	})

	lastEventID, _ = publishTaskEvent(ctx, streamStore, aievent.TaskEvent{
		ProjectID: projectID,
		Type:      aievent.EventAgentStart,
		Agent:     agentName,
		CreatedAt: time.Now().UnixMilli(),
	})
	_ = setProjectState(ctx, stateStore, projectID, aievent.ProjectState{
		Status:      "running",
		Agent:       agentName,
		LastEventID: lastEventID,
		UpdatedAt:   time.Now().UnixMilli(),
	})
	loop.Push(initialMessages)

	controlDone := make(chan struct{})
	go func() {
		defer close(controlDone)
		watchStream(loopCtx, stateStore, streamStore, projectID, loop, &lastEventID)
	}()

	loop.Run(loopCtx)
	state := loop.Wait()
	cancelLoop()
	<-controlDone

	if state != nil && state.ExitReason != nil {
		if state.StopCause != "" {
			lastEventID, _ := publishTaskEvent(ctx, streamStore, aievent.TaskEvent{
				ProjectID: projectID,
				Type:      aievent.EventCancelled,
				Agent:     agentName,
				Content:   state.StopCause,
				CreatedAt: time.Now().UnixMilli(),
			})
			_ = setProjectState(ctx, stateStore, projectID, aievent.ProjectState{
				Status:      "cancelled",
				Agent:       agentName,
				LastEventID: lastEventID,
				Message:     state.StopCause,
				UpdatedAt:   time.Now().UnixMilli(),
			})
			_ = clearProjectState(ctx, stateStore, streamStore, projectID)
			return map[string]any{}, nil
		}
		lastEventID, _ = publishTaskEvent(ctx, streamStore, aievent.TaskEvent{
			ProjectID: projectID,
			Type:      aievent.EventError,
			Agent:     agentName,
			Content:   state.ExitReason.Error(),
			CreatedAt: time.Now().UnixMilli(),
		})
		_ = setProjectState(ctx, stateStore, projectID, aievent.ProjectState{
			Status:      "error",
			Agent:       agentName,
			LastEventID: lastEventID,
			Message:     state.ExitReason.Error(),
			UpdatedAt:   time.Now().UnixMilli(),
		})
		_ = clearProjectState(ctx, stateStore, streamStore, projectID)
		return nil, state.ExitReason
	}

	lastEventID, _ = publishTaskEvent(ctx, streamStore, aievent.TaskEvent{
		ProjectID: projectID,
		Type:      aievent.EventDone,
		Agent:     agentName,
		CreatedAt: time.Now().UnixMilli(),
	})
	_ = setProjectState(ctx, stateStore, projectID, aievent.ProjectState{
		Status:      "done",
		Agent:       agentName,
		LastEventID: lastEventID,
		UpdatedAt:   time.Now().UnixMilli(),
	})
	_ = clearProjectState(ctx, stateStore, streamStore, projectID)
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
