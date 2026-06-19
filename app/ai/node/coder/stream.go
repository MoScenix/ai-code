package coder

import (
	"context"
	"strings"
	"time"

	"github.com/MoScenix/ai-code/app/ai/node/control"
	"github.com/MoScenix/ai-code/app/ai/utils"
	"github.com/MoScenix/ai-code/common/aievent"
	"github.com/MoScenix/ai-code/common/redisstate"
	"github.com/MoScenix/ai-code/common/redisstream"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

func watchStream(ctx context.Context, stateStore *redisstate.Store, store redisstream.Store, projectID string, loop *adk.TurnLoop[[]*schema.Message, *schema.Message], lastEventID *string, assistantOutput *utils.StringBuffer) {
	control.Watch(ctx, store, projectID, controlCursor(ctx), control.Handler{
		OnPush: func(ctx context.Context, msg redisstream.Message, event aievent.TaskEvent) {
			utils.SetControlCursor(ctx, msg.ID)
			handlePush(ctx, stateStore, store, projectID, event.Content, loop, lastEventID, assistantOutput)
		},
		OnCancel: func(ctx context.Context, msg redisstream.Message, event aievent.TaskEvent) {
			utils.SetControlCursor(ctx, msg.ID)
			reason := strings.TrimSpace(event.Content)
			if reason == "" {
				reason = "cancelled"
			}
			utils.CancelRuntime(ctx)
			loop.Stop(adk.WithImmediate(), adk.WithStopCause(reason), adk.WithSkipCheckpoint())
		},
	})
}

func handlePush(ctx context.Context, stateStore *redisstate.Store, store redisstream.Store, projectID string, rawContent string, loop *adk.TurnLoop[[]*schema.Message, *schema.Message], lastEventID *string, assistantOutput *utils.StringBuffer) {
	content := strings.TrimSpace(rawContent)
	if content == "" {
		return
	}
	if buffer, ok := utils.StringBufferFromContext(ctx); ok {
		buffer.Clear()
	}

	accepted, done := loop.Push([]*schema.Message{schema.UserMessage(content)})
	if done != nil {
		go func() { <-done }()
	}
	if !accepted {
		return
	}

	if err := persistPushHistory(ctx, projectID, content, assistantOutput); err != nil {
		setLastEventID(lastEventID, publishError(ctx, store, projectID, err))
		return
	}

	acceptedID := publishAccepted(ctx, store, projectID)
	setLastEventID(lastEventID, acceptedID)
	_ = updateProjectLastEventID(ctx, stateStore, projectID, acceptedID)
}

func persistPushHistory(ctx context.Context, projectID string, content string, assistantOutput *utils.StringBuffer) error {
	if assistantOutput != nil {
		if err := utils.AddProjectAssistantMessage(ctx, projectID, assistantOutput.String()); err != nil {
			return err
		}
	}
	if err := utils.AddProjectUserMessage(ctx, projectID, content); err != nil {
		return err
	}
	if assistantOutput != nil {
		assistantOutput.Clear()
	}
	return nil
}

func publishAccepted(ctx context.Context, store redisstream.Store, projectID string) string {
	id, _ := publishTaskEvent(ctx, store, aievent.TaskEvent{
		ProjectID: projectID,
		Type:      aievent.EventAccepted,
		Agent:     agentName,
		Content:   "push accepted",
		CreatedAt: time.Now().UnixMilli(),
	})
	return id
}

func publishError(ctx context.Context, store redisstream.Store, projectID string, err error) string {
	id, _ := publishTaskEvent(ctx, store, aievent.TaskEvent{
		ProjectID: projectID,
		Type:      aievent.EventError,
		Agent:     agentName,
		Content:   aievent.TrimEventContent(err.Error()),
		CreatedAt: time.Now().UnixMilli(),
	})
	return id
}

func setLastEventID(lastEventID *string, id string) {
	if lastEventID != nil && id != "" {
		*lastEventID = id
	}
}

func controlCursor(ctx context.Context) string {
	if cursor := strings.TrimSpace(utils.ControlCursor(ctx)); cursor != "" {
		return cursor
	}
	return "$"
}
