package service

import (
	"context"
	"time"

	"github.com/MoScenix/ai-code/app/app/biz/dal/mysql"
	"github.com/MoScenix/ai-code/app/app/biz/model"
	app "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
)

type UpdateAppService struct {
	ctx context.Context
} // NewUpdateAppService new UpdateAppService
func NewUpdateAppService(ctx context.Context) *UpdateAppService {
	return &UpdateAppService{ctx: ctx}
}

// Run create note info
func (s *UpdateAppService) Run(req *app.UpdateAppReq) (resp *app.UpdateAppResp, err error) {
	// Finish your business logic.
	q := model.NewAppQuery(s.ctx, mysql.DB)
	up := model.App{
		Name:     req.AppName,
		Cover:    req.Cover,
		Priority: int(req.Priority),
	}
	if req.DeployKey != "" {
		up.Deploykey = req.DeployKey
		up.DeployedTime = time.Now().Format("2006-01-02 15:04:05")
	}
	err = q.UpdateApp(uint(req.Id), up)
	if err != nil {
		return nil, err
	}
	return &app.UpdateAppResp{
		Success: true,
	}, nil
}
