package designer

import (
	"context"
	"strings"
	"time"

	"github.com/MoScenix/ai-code/app/ai/agent"
	"github.com/MoScenix/ai-code/app/ai/utils"
	"github.com/MoScenix/ai-code/common/aievent"
	"github.com/MoScenix/ai-code/common/redisstream"
)

const answerWait = 30 * time.Second

func waitAnswer(ctx context.Context, store redisstream.Store, projectID string, afterID string, targetID string) (agent.DesignerAnswer, bool, error) {
	if store == nil || projectID == "" || targetID == "" {
		return agent.DesignerAnswer{}, false, nil
	}

	deadline := time.Now().Add(answerWait)
	lastID := afterID
	for {
		remain := time.Until(deadline)
		if remain <= 0 {
			return agent.DesignerAnswer{}, false, nil
		}
		messages, err := store.Read(ctx, aievent.StreamKey(projectID), lastID, redisstream.ReadOptions{
			Block: remain,
			Count: 10,
		})
		if err != nil {
			if ctx.Err() != nil {
				return agent.DesignerAnswer{}, false, ctx.Err()
			}
			return agent.DesignerAnswer{}, false, err
		}
		for _, msg := range messages {
			lastID = msg.ID
			event, err := redisstream.Decode[aievent.TaskEvent](msg)
			if err != nil || event.Type != aievent.EventAnswer {
				continue
			}
			if strings.TrimSpace(event.TargetID) != targetID {
				continue
			}
			return agent.DesignerAnswer{
				Content: event.Content,
				Payload: event.Payload,
			}, true, nil
		}
		if ctx.Err() != nil {
			return agent.DesignerAnswer{}, false, ctx.Err()
		}
	}
}

func lastEventCursor(ctx context.Context, projectID string) string {
	stateStore, ok := utils.StateStoreFromContext(ctx)
	if !ok || stateStore == nil || projectID == "" {
		return "$"
	}

	var state aievent.ProjectState
	ok, err := stateStore.Get(ctx, aievent.RunningStateKey(projectID), &state)
	if err != nil || !ok || strings.TrimSpace(state.LastEventID) == "" {
		return "$"
	}
	return state.LastEventID
}
