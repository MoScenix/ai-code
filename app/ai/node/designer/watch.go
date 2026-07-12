package designer

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
	"github.com/cloudwego/kitex/pkg/klog"
)

func watchPushes(ctx context.Context, stateStore *redisstate.Store, store redisstream.Store, projectID string, buffer *utils.StringBuffer, answers chan<- answerEvent, loop *adk.TurnLoop[[]*schema.Message, *schema.Message]) {
	control.Watch(ctx, store, projectID, controlCursor(ctx, projectID), control.Handler{
		OnPush: func(ctx context.Context, msg redisstream.Message, event aievent.TaskEvent) {
			utils.SetControlCursor(ctx, msg.ID)
			content := strings.TrimSpace(event.Content)
			if content != "" && buffer != nil {
				buffer.WriteString(content)
				buffer.WriteString("\n")
			}
			if content != "" && loop != nil {
				_, done := loop.Push([]*schema.Message{schema.UserMessage(content)})
				if done != nil {
					go func() { <-done }()
				}
			}
		},
		OnCancel: func(ctx context.Context, msg redisstream.Message, event aievent.TaskEvent) {
			utils.SetControlCursor(ctx, msg.ID)
			utils.CancelRuntime(ctx)
			if loop != nil {
				reason := strings.TrimSpace(event.Content)
				if reason == "" {
					reason = "cancelled"
				}
				loop.Stop(adk.WithImmediate(), adk.WithStopCause(reason), adk.WithSkipCheckpoint())
			}
		},
		OnAnswer: func(ctx context.Context, msg redisstream.Message, event aievent.TaskEvent) {
			if answers == nil {
				return
			}
			select {
			case answers <- answerEvent{
				TargetID: strings.TrimSpace(event.TargetID),
				Answer:   agentAnswer(event),
			}:
				utils.SetControlCursor(ctx, msg.ID)
			case <-ctx.Done():
				return
			}
			if err := markAnswerAccepted(ctx, projectID, strings.TrimSpace(event.TargetID), msg.ID); err != nil {
				klog.CtxErrorf(ctx, "accept designer answer failed: project_id=%s target_id=%s err=%v", projectID, strings.TrimSpace(event.TargetID), err)
			}
		},
	})
}

func markAnswerAccepted(ctx context.Context, projectID string, targetID string, eventID string) error {
	stateStore, ok := utils.StateStoreFromContext(ctx)
	if !ok || stateStore == nil || projectID == "" || eventID == "" {
		return nil
	}

	var state aievent.ProjectState
	ok, err := stateStore.Get(ctx, aievent.RunningStateKey(projectID), &state)
	if err != nil || !ok || state.Status != aievent.ProjectStatusWaitingAnswer {
		return err
	}
	if targetID != "" && !aievent.PendingInterruptsMatch(state.PendingInterrupts, targetID) {
		return nil
	}

	state.PendingInterrupts = remainingInterrupts(state.PendingInterrupts, targetID)
	if len(state.PendingInterrupts) == 0 {
		state.Status = aievent.ProjectStatusRunning
	}
	state.UpdatedAt = time.Now().UnixMilli()
	return stateStore.Set(ctx, aievent.RunningStateKey(projectID), state)
}

func remainingInterrupts(interrupts []aievent.PendingInterrupt, targetID string) []aievent.PendingInterrupt {
	targetID = strings.TrimSpace(targetID)
	if targetID == "" {
		return interrupts
	}
	out := make([]aievent.PendingInterrupt, 0, len(interrupts))
	for _, interrupt := range interrupts {
		if !aievent.PendingInterruptMatches(interrupt, targetID) {
			out = append(out, interrupt)
		}
	}
	return out
}

func controlCursor(ctx context.Context, projectID string) string {
	if cursor := strings.TrimSpace(utils.ControlCursor(ctx)); cursor != "" {
		return cursor
	}
	return "$"
}
