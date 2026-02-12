package service

import (
	"context"

	"github.com/MoScenix/ai-code/app/app/biz/dal/mysql"
	"github.com/MoScenix/ai-code/app/app/biz/dal/redis"
	"github.com/MoScenix/ai-code/app/app/biz/model"
	app "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
)

type DeleteMessageService struct {
	ctx context.Context
} // NewDeleteMessageService new DeleteMessageService
func NewDeleteMessageService(ctx context.Context) *DeleteMessageService {
	return &DeleteMessageService{ctx: ctx}
}

// Run create note info
func (s *DeleteMessageService) Run(req *app.DeleteMessageReq) (resp *app.DeleteMessageResp, err error) {
	op, err := getOperator(s.ctx)
	if err != nil {
		return nil, err
	}
	msgQuery := model.NewMessageQuery(s.ctx, mysql.DB)
	msg, err := msgQuery.GetMessageById(uint(req.Id))
	if err != nil {
		return nil, err
	}
	appInfo, err := model.NewAppProQuery(s.ctx, mysql.DB, redis.RedisClient).GetAppById(msg.AppId)
	if err != nil {
		return nil, err
	}
	if err = mustOwnerOrAdmin(op, appInfo.UserId); err != nil {
		return nil, err
	}
	err = msgQuery.DeleteMessageById(uint(req.Id))
	if err != nil {
		return nil, err
	}
	return &app.DeleteMessageResp{
		Success: true,
	}, nil
}
