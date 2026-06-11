package designer

import (
	"context"
	"strings"
	"time"

	"github.com/MoScenix/ai-code/app/ai/utils"
	"github.com/MoScenix/ai-code/common/aievent"
	"github.com/MoScenix/ai-code/common/redisstream"
)

func watchPushes(ctx context.Context, store redisstream.Store, projectID string, lastEventID *string, buffer *utils.StringBuffer) {
	if store == nil || projectID == "" {
		return
	}

	lastID := lastEventCursor(ctx, projectID)
	for {
		if ctx.Err() != nil {
			return
		}
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
			if lastEventID != nil {
				*lastEventID = msg.ID
			}

			event, err := redisstream.Decode[aievent.TaskEvent](msg)
			if err != nil {
				continue
			}
			switch event.Type {
			case aievent.EventPush:
				content := strings.TrimSpace(event.Content)
				if content != "" && buffer != nil {
					buffer.WriteString(content)
					buffer.WriteString("\n")
				}
			case aievent.EventCancel:
				utils.CancelRuntime(ctx)
			}
		}
	}
}
