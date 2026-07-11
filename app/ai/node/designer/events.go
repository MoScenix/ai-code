package designer

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/MoScenix/ai-code/app/ai/agent"
	"github.com/MoScenix/ai-code/app/ai/utils"
	"github.com/MoScenix/ai-code/common/aievent"
	"github.com/MoScenix/ai-code/common/redisstate"
	"github.com/MoScenix/ai-code/common/redisstream"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/kitex/pkg/klog"
)

func publishAgentEvents(ctx context.Context, store redisstream.Store, projectID string, events *adk.AsyncIterator[*adk.TypedAgentEvent[*schema.Message]], lastID *string) (*interruptEvent, string, error) {
	var content strings.Builder
	for {
		event, ok := events.Next()
		if !ok {
			return nil, content.String(), nil
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
			updateLastID(lastID, id)
			continue
		}
		if event.Action != nil && event.Action.Interrupted != nil {
			interrupt := newInterruptEvent(event.AgentName, event.Action.Interrupted)
			id, _ := publishTaskEvent(ctx, store, aievent.TaskEvent{
				ProjectID: projectID,
				Type:      aievent.EventQuestion,
				Agent:     event.AgentName,
				TargetID:  interrupt.ID,
				Content:   interrupt.Content,
				Payload:   interrupt.Payload,
				CreatedAt: time.Now().UnixMilli(),
			})
			updateLastID(lastID, id)
			utils.SetControlCursor(ctx, id)
			interrupt.EventID = id
			if stateStore, ok := utils.StateStoreFromContext(ctx); ok && stateStore != nil {
				_ = setProjectState(ctx, stateStore, projectID, aievent.ProjectState{
					Status:      aievent.ProjectStatusWaitingAnswer,
					Agent:       event.AgentName,
					LastEventID: stateLastEventID(ctx, stateStore, projectID),
					PendingInterrupts: []aievent.PendingInterrupt{
						{
							ID:      interrupt.ID,
							Agent:   event.AgentName,
							Content: interrupt.Content,
							Payload: interrupt.Payload,
						},
					},
					UpdatedAt: time.Now().UnixMilli(),
				})
			}
			return interrupt, content.String(), nil
		}
		if event.Output == nil || event.Output.MessageOutput == nil {
			continue
		}
		id, text, err := publishMessageOutput(ctx, store, projectID, event.AgentName, event.Output.MessageOutput)
		if err != nil {
			return nil, content.String(), err
		}
		content.WriteString(text)
		updateLastID(lastID, id)
	}
}

func publishMessageOutput(ctx context.Context, store redisstream.Store, projectID string, agentName string, output *adk.TypedMessageVariant[*schema.Message]) (string, string, error) {
	if output.IsStreaming {
		var lastID string
		var content strings.Builder
		for {
			msg, err := output.MessageStream.Recv()
			if err != nil {
				if err == io.EOF {
					return lastID, content.String(), nil
				}
				return lastID, content.String(), err
			}
			id, _ := publishSchemaMessage(ctx, store, projectID, agentName, output.Role, output.ToolName, msg)
			if id != "" {
				lastID = id
			}
			if msg != nil {
				content.WriteString(msg.Content)
			}
		}
	}
	msg, err := output.GetMessage()
	if err != nil {
		return "", "", err
	}
	id, err := publishSchemaMessage(ctx, store, projectID, agentName, output.Role, output.ToolName, msg)
	if msg == nil {
		return id, "", err
	}
	return id, msg.Content, err
}

func publishSchemaMessage(ctx context.Context, store redisstream.Store, projectID string, agentName string, role schema.RoleType, toolName string, msg *schema.Message) (string, error) {
	if msg == nil {
		return "", nil
	}
	eventType := aievent.EventMessage
	content := msg.Content
	if role == schema.Tool {
		eventType = aievent.EventToolResult
		content = aievent.TrimEventContent(content)
	}
	return publishTaskEvent(ctx, store, aievent.TaskEvent{
		ProjectID: projectID,
		Type:      eventType,
		Agent:     agentName,
		Content:   content,
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

func setProjectState(ctx context.Context, store *redisstate.Store, projectID string, state aievent.ProjectState) error {
	if store == nil || projectID == "" {
		return nil
	}
	return store.Set(ctx, aievent.RunningStateKey(projectID), state)
}

func stateLastEventID(ctx context.Context, store *redisstate.Store, projectID string) string {
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

func updateLastID(lastID *string, id string) {
	if lastID != nil && id != "" {
		*lastID = id
	}
}

type interruptEvent struct {
	ID      string
	EventID string
	Content string
	Payload map[string]any
}

func newInterruptEvent(agentName string, info *adk.InterruptInfo) *interruptEvent {
	userInfo := interruptUserInfo(info)
	payload := map[string]any{}
	id := rootInterruptID(info)
	content := fmt.Sprint(userInfo)
	if input, ok := userInfo.(agent.AskUserInput); ok {
		questions := normalizeAskQuestions(input.Questions)
		content = askQuestionsContent(questions)
		payload["questions"] = questions
		payload["context"] = input.Context
	}
	payload["agent"] = agentName
	return &interruptEvent{
		ID:      id,
		Content: content,
		Payload: payload,
	}
}

func normalizeAskQuestions(questions []agent.AskUserQuestion) []agent.AskUserQuestion {
	out := make([]agent.AskUserQuestion, 0, len(questions))
	for _, question := range questions {
		text := strings.TrimSpace(question.Question)
		if text == "" {
			continue
		}
		options := make([]string, 0, len(question.Options))
		for _, option := range question.Options {
			option = strings.TrimSpace(option)
			if option != "" {
				options = append(options, option)
			}
		}
		out = append(out, agent.AskUserQuestion{
			Question: text,
			Options:  options,
		})
	}
	return out
}

func askQuestionsContent(questions []agent.AskUserQuestion) string {
	lines := make([]string, 0, len(questions))
	for _, question := range questions {
		lines = append(lines, question.Question)
	}
	return strings.Join(lines, "\n")
}

func interruptUserInfo(info *adk.InterruptInfo) any {
	for _, ctx := range info.InterruptContexts {
		if ctx != nil && ctx.IsRootCause && ctx.Info != nil {
			return ctx.Info
		}
	}
	for _, ctx := range info.InterruptContexts {
		if ctx != nil && ctx.Info != nil {
			return ctx.Info
		}
	}
	if cmInfo, ok := info.Data.(*adk.ChatModelAgentInterruptInfo); ok && cmInfo != nil && cmInfo.Info != nil {
		for _, ctx := range cmInfo.Info.InterruptContexts {
			if ctx != nil && ctx.IsRootCause && ctx.Info != nil {
				return ctx.Info
			}
		}
		for _, ctx := range cmInfo.Info.InterruptContexts {
			if ctx != nil && ctx.Info != nil {
				return ctx.Info
			}
		}
	}
	return info.Data
}

func rootInterruptID(info *adk.InterruptInfo) string {
	for _, ctx := range info.InterruptContexts {
		if ctx != nil && ctx.IsRootCause && strings.TrimSpace(ctx.ID) != "" {
			return ctx.ID
		}
	}
	for _, ctx := range info.InterruptContexts {
		if ctx != nil && strings.TrimSpace(ctx.ID) != "" {
			return ctx.ID
		}
	}
	if cmInfo, ok := info.Data.(*adk.ChatModelAgentInterruptInfo); ok && cmInfo != nil && cmInfo.Info != nil {
		for _, ctx := range cmInfo.Info.InterruptContexts {
			if ctx != nil && ctx.IsRootCause && strings.TrimSpace(ctx.ID) != "" {
				return ctx.ID
			}
		}
		for _, ctx := range cmInfo.Info.InterruptContexts {
			if ctx != nil && strings.TrimSpace(ctx.ID) != "" {
				return ctx.ID
			}
		}
	}
	return ""
}
