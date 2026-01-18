package main

import (
	"context"

	"github.com/MoScenix/ai-code/app/ai/biz/service"
	ai "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/ai"
)

// AiServiceImpl implements the last service interface defined in the IDL.
type AiServiceImpl struct{}

func (s *AiServiceImpl) Chat(req *ai.AiReq, stream ai.AiService_ChatServer) (err error) {
	ctx := context.Background()
	err = service.NewChatService(ctx).Run(req, stream)
	return
}
