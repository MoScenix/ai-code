package service

import (
	"context"
	"os"
	"path/filepath"
	"strconv"

	"github.com/MoScenix/ai-code/app/app/biz/dal/mysql"
	"github.com/MoScenix/ai-code/app/app/biz/dal/redis"
	"github.com/MoScenix/ai-code/app/app/biz/model"
	"github.com/MoScenix/ai-code/app/app/conf"
	app "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
	"github.com/cloudwego/kitex/pkg/klog"
)

type AddAppService struct {
	ctx context.Context
} // NewAddAppService new AddAppService
func NewAddAppService(ctx context.Context) *AddAppService {
	return &AddAppService{ctx: ctx}
}

// Run create note info
func (s *AddAppService) Run(req *app.AddAppReq) (resp *app.AddAppResp, err error) {
	if mysql.DB == nil {
		return nil, errDBNotReady
	}
	op, err := getOperator(s.ctx)
	if err != nil {
		klog.CtxErrorf(s.ctx, "create app operator check failed: err=%v", err)
		return nil, err
	}
	if req.UserId != 0 && !op.isAdmin() && uint(req.UserId) != op.userID {
		klog.CtxWarnf(s.ctx, "create app forbidden: user_id=%d operator_id=%d", req.UserId, op.userID)
		return nil, errForbidden
	}
	if req.UserId == 0 || !op.isAdmin() {
		req.UserId = int64(op.userID)
	}
	rs := []rune(req.InitPrompt)
	res, err := model.NewAppProQuery(s.ctx, mysql.DB, redis.RedisClient).CreateApp(model.App{
		Name:       string(rs[:min(len(rs), 12)]),
		InitPrompt: req.InitPrompt,
		UserId:     uint(req.UserId),
		Priority:   1,
	})
	if err != nil {
		klog.CtxErrorf(s.ctx, "create app db failed: user_id=%d err=%v", req.UserId, err)
		return nil, err
	}
	path := filepath.Join(conf.GetConf().ShareDir.ShareDir, strconv.FormatInt(int64(res.ID), 10))
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		klog.CtxErrorf(s.ctx, "create app directory failed: app_id=%d err=%v", res.ID, err)
		return nil, err
	}
	klog.CtxInfof(s.ctx, "app created: app_id=%d user_id=%d", res.ID, req.UserId)
	return &app.AddAppResp{
		Id: int64(res.ID),
	}, err
}
