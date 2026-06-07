package service

import (
	"context"
	"errors"

	ai "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/ai"
)

var ErrChatRuntimeUnavailable = errors.New("ai chat runtime is under reconstruction")

type ChatService struct {
	ctx context.Context
}

// NewChatService new ChatService
func NewChatService(ctx context.Context) *ChatService {
	return &ChatService{ctx: ctx}
}

func (s *ChatService) Run(req *ai.AiReq, stream ai.AiService_ChatServer) (err error) {
	return ErrChatRuntimeUnavailable
}
