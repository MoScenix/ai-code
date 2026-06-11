package coder

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/MoScenix/ai-code/app/ai/utils"
	"github.com/MoScenix/ai-code/common/aievent"
	"github.com/MoScenix/ai-code/common/redisstate"
	"github.com/MoScenix/ai-code/common/redisstream"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

func publishAgentEvents(ctx context.Context, store redisstream.Store, projectID string, events *adk.AsyncIterator[*adk.TypedAgentEvent[*schema.Message]], lastID *string) error {
	for {
		event, ok := events.Next()
		if !ok {
			return nil
		}
		if event == nil {
			continue
		}
		if event.Err != nil {
			id, _ := publishTaskEvent(ctx, store, aievent.TaskEvent{
				ProjectID: projectID,
				Type:      aievent.EventError,
				Agent:     event.AgentName,
				Content:   event.Err.Error(),
				CreatedAt: time.Now().UnixMilli(),
			})
			if lastID != nil && id != "" {
				*lastID = id
			}
			continue
		}
		if event.Action != nil && event.Action.Interrupted != nil {
			id, _ := publishTaskEvent(ctx, store, aievent.TaskEvent{
				ProjectID: projectID,
				Type:      aievent.EventQuestion,
				Agent:     event.AgentName,
				Content:   fmt.Sprint(event.Action.Interrupted.Data),
				Payload: map[string]any{
					"interrupt_contexts": event.Action.Interrupted.InterruptContexts,
				},
				CreatedAt: time.Now().UnixMilli(),
			})
			if lastID != nil && id != "" {
				*lastID = id
			}
			continue
		}
		if event.Output == nil || event.Output.MessageOutput == nil {
			continue
		}
		id, err := publishMessageOutput(ctx, store, projectID, event.AgentName, event.Output.MessageOutput)
		if err != nil {
			return err
		}
		if lastID != nil && id != "" {
			*lastID = id
		}
	}
}

func publishMessageOutput(ctx context.Context, store redisstream.Store, projectID string, agentName string, output *adk.TypedMessageVariant[*schema.Message]) (string, error) {
	if output.IsStreaming {
		var lastID string
		for {
			msg, err := output.MessageStream.Recv()
			if err != nil {
				if err == io.EOF {
					return lastID, nil
				}
				return lastID, err
			}
			id, _ := publishSchemaMessage(ctx, store, projectID, agentName, output.Role, output.ToolName, msg)
			if id != "" {
				lastID = id
			}
		}
	}
	msg, err := output.GetMessage()
	if err != nil {
		return "", err
	}
	return publishSchemaMessage(ctx, store, projectID, agentName, output.Role, output.ToolName, msg)
}

func publishSchemaMessage(ctx context.Context, store redisstream.Store, projectID string, agentName string, role schema.RoleType, toolName string, msg *schema.Message) (string, error) {
	if msg == nil {
		return "", nil
	}
	eventType := aievent.EventMessage
	if role == schema.Tool {
		eventType = aievent.EventToolResult
	}
	return publishTaskEvent(ctx, store, aievent.TaskEvent{
		ProjectID: projectID,
		Type:      eventType,
		Agent:     agentName,
		Content:   msg.Content,
		Name:      toolName,
		CreatedAt: time.Now().UnixMilli(),
	})
}

func publishTaskEvent(ctx context.Context, store redisstream.Store, event aievent.TaskEvent) (string, error) {
	if store == nil || event.ProjectID == "" {
		return "", nil
	}
	if event.CreatedAt == 0 {
		event.CreatedAt = time.Now().UnixMilli()
	}
	id, err := store.Add(ctx, aievent.StreamKey(event.ProjectID), event)
	if err != nil {
		return "", err
	}
	if stateStore, ok := utils.StateStoreFromContext(ctx); ok && stateStore != nil {
		_ = stateStore.Set(ctx, aievent.RunningStateKey(event.ProjectID), aievent.ProjectState{
			Status:      "running",
			LastEventID: id,
			UpdatedAt:   time.Now().UnixMilli(),
		})
	}
	return id, nil
}

func setProjectState(ctx context.Context, store *redisstate.Store, projectID string, state aievent.ProjectState) error {
	if store == nil || projectID == "" {
		return nil
	}
	return store.Set(ctx, aievent.RunningStateKey(projectID), state)
}

func clearProjectState(ctx context.Context, stateStore *redisstate.Store, streamStore redisstream.Store, projectID string) error {
	if projectID == "" {
		return nil
	}
	if stateStore != nil {
		_ = stateStore.Del(ctx, aievent.RunningStateKey(projectID), aievent.CursorKey(projectID), aievent.ActiveTaskKey(projectID))
	}
	if streamStore != nil {
		_ = streamStore.Del(ctx, aievent.StreamKey(projectID))
	}
	return nil
}
