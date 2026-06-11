package service

import (
	"context"

	aitask "github.com/MoScenix/ai-code/app/ai/task"
	aiworkpool "github.com/MoScenix/ai-code/app/ai/workpool"
	ai "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/ai"
)

type ChatService struct {
	ctx context.Context
}

// NewChatService new ChatService
func NewChatService(ctx context.Context) *ChatService {
	return &ChatService{ctx: ctx}
}

func (s *ChatService) Run(projectID string, stream ai.AiService_ChatServer) (err error) {
	runCtx := context.WithoutCancel(s.ctx)
	task := aitask.NewChatTask(projectID, stream)
	if err := task.Enqueue(runCtx); err != nil {
		_ = sendSubmitResult(stream, false)
		return err
	}

	p, err := aiworkpool.Get()
	if err != nil {
		_ = sendSubmitResult(stream, false)
		return err
	}
	if err := p.Submit(runCtx, task); err != nil {
		_ = sendSubmitResult(stream, false)
		return err
	}
	return sendSubmitResult(stream, true)
}

func sendSubmitResult(stream ai.AiService_ChatServer, ok bool) error {
	if stream == nil {
		return nil
	}
	answer := "false"
	if ok {
		answer = "true"
	}
	return stream.Send(&ai.AiResp{Answer: answer})
}
