package coder

import (
	"context"
	"strings"
	"time"

	"github.com/MoScenix/ai-code/app/ai/utils"
	"github.com/MoScenix/ai-code/common/aievent"
	"github.com/MoScenix/ai-code/common/redisstate"
	"github.com/MoScenix/ai-code/common/redisstream"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

func watchStream(ctx context.Context, stateStore *redisstate.Store, store redisstream.Store, projectID string, loop *adk.TurnLoop[[]*schema.Message, *schema.Message], lastEventID *string) {
	if store == nil || projectID == "" {
		return
	}

	lastID := lastEventCursor(ctx, stateStore, projectID)
	for {
		messages, err := store.Read(ctx, aievent.StreamKey(projectID), lastID, redisstream.ReadOptions{
			Block: 30 * time.Second,
			Count: 10,
		})
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			continue
		}

		for _, msg := range messages {
			lastID = msg.ID
			event, err := redisstream.Decode[aievent.TaskEvent](msg)
			if err != nil {
				continue
			}

			switch event.Type {
			case aievent.EventPush:
				content := strings.TrimSpace(event.Content)
				if content == "" {
					continue
				}
				if buffer, ok := utils.StringBufferFromContext(ctx); ok {
					buffer.Clear()
				}
				accepted, done := loop.Push([]*schema.Message{schema.UserMessage(content)}, adk.WithPreempt[[]*schema.Message, *schema.Message](adk.AfterChatModel))
				if accepted {
					id, _ := publishTaskEvent(ctx, store, aievent.TaskEvent{
						ProjectID: projectID,
						Type:      aievent.EventAccepted,
						Agent:     agentName,
						Content:   "push accepted",
						CreatedAt: time.Now().UnixMilli(),
					})
					if lastEventID != nil && id != "" {
						*lastEventID = id
					}
				}
				if done != nil {
					go func() { <-done }()
				}
			case aievent.EventCancel:
				reason := strings.TrimSpace(event.Content)
				if reason == "" {
					reason = "cancelled"
				}
				utils.CancelRuntime(ctx)
				loop.Stop(adk.WithImmediate(), adk.WithStopCause(reason), adk.WithSkipCheckpoint())
			}
		}

		if ctx.Err() != nil {
			return
		}
	}
}

func lastEventCursor(ctx context.Context, stateStore *redisstate.Store, projectID string) string {
	if stateStore == nil || projectID == "" {
		return "$"
	}

	var state aievent.ProjectState
	ok, err := stateStore.Get(ctx, aievent.RunningStateKey(projectID), &state)
	if err != nil || !ok || strings.TrimSpace(state.LastEventID) == "" {
		return "$"
	}
	return state.LastEventID
}
