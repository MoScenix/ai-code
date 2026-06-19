package control

import (
	"context"
	"strings"
	"time"

	"github.com/MoScenix/ai-code/common/aievent"
	"github.com/MoScenix/ai-code/common/redisstream"
)

type Handler struct {
	OnPush   func(ctx context.Context, msg redisstream.Message, event aievent.TaskEvent)
	OnCancel func(ctx context.Context, msg redisstream.Message, event aievent.TaskEvent)
	OnAnswer func(ctx context.Context, msg redisstream.Message, event aievent.TaskEvent)
}

func Watch(ctx context.Context, store redisstream.Store, projectID string, cursor string, handler Handler) {
	if store == nil || projectID == "" {
		return
	}
	if strings.TrimSpace(cursor) == "" {
		cursor = "$"
	}

	lastID := cursor
	for {
		if ctx.Err() != nil {
			return
		}
		messages, err := store.Read(ctx, aievent.ControlKey(projectID), lastID, redisstream.ReadOptions{
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
				if handler.OnPush != nil {
					handler.OnPush(ctx, msg, event)
				}
			case aievent.EventCancel:
				if handler.OnCancel != nil {
					handler.OnCancel(ctx, msg, event)
				}
			case aievent.EventAnswer:
				if handler.OnAnswer != nil {
					handler.OnAnswer(ctx, msg, event)
				}
			}
		}
	}
}
