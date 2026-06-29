package coder

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/MoScenix/ai-code/app/ai/utils"
	"github.com/MoScenix/ai-code/common/aievent"
	"github.com/MoScenix/ai-code/common/redisstate"
	"github.com/MoScenix/ai-code/common/redisstream"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/kitex/pkg/klog"
)

func publishAgentEvents(ctx context.Context, store redisstream.Store, projectID string, events *adk.AsyncIterator[*adk.TypedAgentEvent[*schema.Message]], lastID *string, assistantOutput *utils.StringBuffer) error {
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
				Content:   aievent.TrimEventContent(event.Err.Error()),
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
		id, err := publishMessageOutput(ctx, store, projectID, event.AgentName, event.Output.MessageOutput, assistantOutput)
		if err != nil {
			return err
		}
		if lastID != nil && id != "" {
			*lastID = id
		}
	}
}

func publishMessageOutput(ctx context.Context, store redisstream.Store, projectID string, agentName string, output *adk.TypedMessageVariant[*schema.Message], assistantOutput *utils.StringBuffer) (string, error) {
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
			id, _ := publishSchemaMessage(ctx, store, projectID, agentName, output.Role, output.ToolName, msg, assistantOutput)
			if id != "" {
				lastID = id
			}
		}
	}
	msg, err := output.GetMessage()
	if err != nil {
		return "", err
	}
	return publishSchemaMessage(ctx, store, projectID, agentName, output.Role, output.ToolName, msg, assistantOutput)
}

func publishSchemaMessage(ctx context.Context, store redisstream.Store, projectID string, agentName string, role schema.RoleType, toolName string, msg *schema.Message, assistantOutput *utils.StringBuffer) (string, error) {
	if msg == nil {
		return "", nil
	}
	effectiveRole := role
	if effectiveRole == "" {
		effectiveRole = msg.Role
	}
	var lastID string
	if effectiveRole == schema.Assistant && len(msg.ToolCalls) > 0 {
		for _, toolCall := range msg.ToolCalls {
			name := strings.TrimSpace(toolCall.Function.Name)
			if name == "" {
				name = strings.TrimSpace(toolName)
			}
			id, err := publishTaskEvent(ctx, store, aievent.TaskEvent{
				ProjectID: projectID,
				Type:      aievent.EventToolCall,
				Agent:     agentName,
				TargetID:  toolCall.ID,
				Name:      name,
				Payload: map[string]any{
					"arguments": toolCall.Function.Arguments,
					"type":      toolCall.Type,
				},
				CreatedAt: time.Now().UnixMilli(),
			})
			if err != nil {
				return lastID, err
			}
			if id != "" {
				lastID = id
			}
		}
	}

	eventType := aievent.EventMessage
	content := messageText(msg)
	if effectiveRole == schema.Tool {
		eventType = aievent.EventToolResult
		content = aievent.TrimEventContent(content)
		if toolName == "" {
			toolName = msg.ToolName
		}
	} else if effectiveRole == schema.Assistant && assistantOutput != nil {
		assistantOutput.WriteString(content)
	}
	if strings.TrimSpace(content) == "" && effectiveRole != schema.Tool {
		return lastID, nil
	}
	id, err := publishTaskEvent(ctx, store, aievent.TaskEvent{
		ProjectID: projectID,
		Type:      eventType,
		Agent:     agentName,
		Content:   content,
		TargetID:  msg.ToolCallID,
		Name:      toolName,
		CreatedAt: time.Now().UnixMilli(),
	})
	if id != "" {
		lastID = id
	}
	return lastID, err
}

func messageText(msg *schema.Message) string {
	if msg == nil {
		return ""
	}
	if msg.Content != "" {
		return msg.Content
	}
	if len(msg.AssistantGenMultiContent) == 0 {
		return ""
	}

	var b strings.Builder
	for _, part := range msg.AssistantGenMultiContent {
		if part.Text != "" {
			b.WriteString(part.Text)
		}
	}
	return b.String()
}

func publishTaskEvent(ctx context.Context, store redisstream.Store, event aievent.TaskEvent) (string, error) {
	if store == nil || event.ProjectID == "" {
		return "", nil
	}
	if event.CreatedAt == 0 {
		event.CreatedAt = time.Now().UnixMilli()
	}
	if event.Type == aievent.EventError {
		klog.CtxErrorf(ctx, "publish ai error event: project_id=%s agent=%s content=%s", event.ProjectID, event.Agent, event.Content)
	}
	id, err := store.Add(ctx, aievent.EventKey(event.ProjectID), event)
	if err != nil {
		if event.Type == aievent.EventError {
			klog.CtxErrorf(ctx, "publish ai error event failed: project_id=%s agent=%s err=%v", event.ProjectID, event.Agent, err)
		}
		return "", err
	}
	return id, nil
}

func publishTerminalEvent(ctx context.Context, streamStore redisstream.Store, stateStore *redisstate.Store, projectID string, eventType aievent.EventType, status string, message string) {
	_, _ = publishTaskEvent(ctx, streamStore, aievent.TaskEvent{
		ProjectID: projectID,
		Type:      eventType,
		Agent:     agentName,
		Content:   terminalContent(eventType, message),
		CreatedAt: time.Now().UnixMilli(),
	})
	_ = setProjectState(ctx, stateStore, projectID, aievent.ProjectState{
		Status:      status,
		Agent:       agentName,
		LastEventID: projectLastEventID(ctx, stateStore, projectID),
		Message:     message,
		UpdatedAt:   time.Now().UnixMilli(),
	})
	_ = expireTerminalTask(ctx, stateStore, streamStore, projectID, terminalTaskTTL)
}

func terminalContent(eventType aievent.EventType, message string) string {
	if eventType == aievent.EventError {
		return aievent.TrimEventContent(message)
	}
	if eventType == aievent.EventCancelled {
		return message
	}
	return ""
}

func setProjectState(ctx context.Context, store *redisstate.Store, projectID string, state aievent.ProjectState) error {
	if store == nil || projectID == "" {
		return nil
	}
	return store.Set(ctx, aievent.RunningStateKey(projectID), state)
}

func projectLastEventID(ctx context.Context, store *redisstate.Store, projectID string) string {
	if store == nil || projectID == "" {
		return ""
	}
	var state aievent.ProjectState
	ok, err := store.Get(ctx, aievent.RunningStateKey(projectID), &state)
	if err != nil || !ok {
		return ""
	}
	return state.LastEventID
}

func updateProjectLastEventID(ctx context.Context, store *redisstate.Store, projectID string, lastEventID string) error {
	if store == nil || projectID == "" || lastEventID == "" {
		return nil
	}
	var state aievent.ProjectState
	ok, err := store.Get(ctx, aievent.RunningStateKey(projectID), &state)
	if err != nil {
		return err
	}
	if !ok {
		state.Status = aievent.ProjectStatusRunning
	}
	state.LastEventID = lastEventID
	state.UpdatedAt = time.Now().UnixMilli()
	return setProjectState(ctx, store, projectID, state)
}

func expireTerminalTask(ctx context.Context, stateStore *redisstate.Store, streamStore redisstream.Store, projectID string, ttl time.Duration) error {
	if projectID == "" {
		return nil
	}
	if stateStore != nil {
		_ = stateStore.Del(ctx, aievent.ActiveTaskKey(projectID))
		_ = stateStore.Expire(ctx, aievent.RunningStateKey(projectID), ttl)
		_ = stateStore.Expire(ctx, aievent.CursorKey(projectID), ttl)
		_ = stateStore.Expire(ctx, aievent.CheckpointKey(projectID), ttl)
		_ = stateStore.Expire(ctx, aievent.GraphCheckpointKey(projectID), ttl)
	}
	if streamStore != nil {
		_ = streamStore.Expire(ctx, aievent.EventKey(projectID), ttl)
		_ = streamStore.Expire(ctx, aievent.ControlKey(projectID), ttl)
	}
	return nil
}
