package main

import (
	"context"

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

func (s *AiServiceImpl) Chat(ctx context.Context, req *ai.AiReq) (resp *ai.AiResp, err error) {
	ctx, err = middleware.InjectHistory(ctx, req.GetProjectId())
	if err != nil {
		return nil, err
	}
	streamStore, err := redisstream.NewRedisStore(redis.RedisClient, "ai")
	if err != nil {
		return nil, err
	}
	stateStore, err := redisstate.NewStore(redis.RedisClient, "ai")
	if err != nil {
		return nil, err
	}
	ctx = utils.WithStreamStore(ctx, streamStore)
	ctx = utils.WithStateStore(ctx, stateStore)
	ok, err := service.NewChatService(ctx).Run(req.GetProjectId())
	if err != nil {
		return nil, err
	}
	answer := "false"
	if ok {
		answer = "true"
	}
	return &ai.AiResp{Answer: answer}, nil
}
