package main

import (
	"github.com/MoScenix/ai-code/app/ai/biz/dal/redis"
	"github.com/MoScenix/ai-code/app/ai/biz/service"
	"github.com/MoScenix/ai-code/app/ai/middleware"
	"github.com/MoScenix/ai-code/app/ai/utils"
	"github.com/MoScenix/ai-code/common/redisstate"
	"github.com/MoScenix/ai-code/common/redisstream"
	ai "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/ai"
)

// AiServiceImpl implements the last service interface defined in the IDL.
type AiServiceImpl struct{}

func (s *AiServiceImpl) Chat(req *ai.AiReq, stream ai.AiService_ChatServer) (err error) {
	ctx, err := middleware.InjectHistory(stream.Context(), req.GetProjectId())
	if err != nil {
		return err
	}
	streamStore, err := redisstream.NewRedisStore(redis.RedisClient, "ai")
	if err != nil {
		return err
	}
	stateStore, err := redisstate.NewStore(redis.RedisClient, "ai")
	if err != nil {
		return err
	}
	ctx = utils.WithStreamStore(ctx, streamStore)
	ctx = utils.WithStateStore(ctx, stateStore)
	err = service.NewChatService(ctx).Run(req.GetProjectId(), stream)
	return
}
