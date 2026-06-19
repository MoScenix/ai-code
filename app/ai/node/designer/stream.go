package designer

import (
	"context"
	"time"

	"github.com/MoScenix/ai-code/app/ai/agent"
	"github.com/MoScenix/ai-code/common/aievent"
)

const answerWait = 60 * time.Second

type answerEvent struct {
	TargetID string
	Answer   agent.DesignerAnswer
}

func waitAnswer(ctx context.Context, answers <-chan answerEvent, targetID string) (agent.DesignerAnswer, bool, error) {
	if answers == nil || targetID == "" {
		return agent.DesignerAnswer{}, false, nil
	}

	timer := time.NewTimer(answerWait)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return agent.DesignerAnswer{}, false, ctx.Err()
		case <-timer.C:
			return agent.DesignerAnswer{}, false, nil
		case event := <-answers:
			if event.TargetID != "" && event.TargetID != targetID {
				continue
			}
			return event.Answer, true, nil
		}
	}
}

func agentAnswer(event aievent.TaskEvent) agent.DesignerAnswer {
	return agent.DesignerAnswer{
		Content: event.Content,
		Payload: event.Payload,
	}
}
