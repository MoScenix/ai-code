package service

import (
	"context"

	"github.com/MoScenix/ai-code/app/app/biz/dal/mysql"
	"github.com/MoScenix/ai-code/app/app/biz/dal/redis"
	"github.com/MoScenix/ai-code/app/app/biz/model"
	app "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
)

type AddMessageService struct {
	ctx context.Context
} // NewAddMessageService new AddMessageService
func NewAddMessageService(ctx context.Context) *AddMessageService {
	return &AddMessageService{ctx: ctx}
}

// Run create note info
func (s *AddMessageService) Run(req *app.AddMessageReq) (resp *app.AddMessageResp, err error) {
	op, err := getOperator(s.ctx)
	if err != nil {
		return nil, err
	}
	if req.UserId != 0 && !op.isAdmin() && uint(req.UserId) != op.userID {
		return nil, errForbidden
	}
	appInfo, err := model.NewAppProQuery(s.ctx, mysql.DB, redis.RedisClient).GetAppById(uint(req.AppId))
	if err != nil {
		return nil, err
	}
	if err = mustOwnerOrAdmin(op, appInfo.UserId); err != nil {
		return nil, err
	}
	res, err := model.NewMessageQuery(s.ctx, mysql.DB).CreateMessage(model.Message{
		AppId:   uint(req.AppId),
		Role:    req.Role,
		Content: req.Content,
	})
	if err != nil {
		return nil, err
	}
	return &app.AddMessageResp{
		Id: int64(res.ID),
	}, nil
}
