package service

import (
	"context"

	"github.com/MoScenix/ai-code/app/app/biz/dal/mysql"
	"github.com/MoScenix/ai-code/app/app/biz/model"
	app "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
)

type GetAppService struct {
	ctx context.Context
} // NewGetAppService new GetAppService
func NewGetAppService(ctx context.Context) *GetAppService {
	return &GetAppService{ctx: ctx}
}

// Run create note info
func (s *GetAppService) Run(req *app.GetAppReq) (resp *app.GetAppResp, err error) {
	// Finish your business logic.
	res, err := model.NewAppQuery(s.ctx, mysql.DB).GetAppById(uint(req.Id))
	if err != nil {
		return nil, err
	}
	return &app.GetAppResp{
		App: &app.AppInfo{
			Id:           int64(res.ID),
			AppName:      res.Name,
			InitPrompt:   res.InitPrompt,
			Cover:        res.Cover,
			DeployKey:    res.Deploykey,
			DeployedTime: res.DeployedTime,
			Priority:     int64(res.Priority),
			UserId:       int64(res.UserId),
			UpdateTime:   res.UpdatedAt.Format("2006-01-02 15:04:05"),
			CreateTime:   res.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	}, nil
}
