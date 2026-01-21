package service

import (
	"context"

	"github.com/MoScenix/ai-code/app/app/biz/dal/mysql"
	"github.com/MoScenix/ai-code/app/app/biz/model"
	app "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
)

type ListAppService struct {
	ctx context.Context
} // NewListAppService new ListAppService
func NewListAppService(ctx context.Context) *ListAppService {
	return &ListAppService{ctx: ctx}
}

// Run create note info
func (s *ListAppService) Run(req *app.ListAppReq) (resp *app.ListAppResp, err error) {
	// Finish your business logic.
	q := model.NewAppQuery(s.ctx, mysql.DB)
	res, err := q.ListApp(uint32(req.PageNum), uint(req.UserId), req.AppName, uint32(req.PageSize))
	if err != nil {
		return nil, err
	}
	tot, err := q.CountApp(uint(req.UserId), req.AppName)
	if err != nil {
		return nil, err
	}
	resp = &app.ListAppResp{
		Total: tot,
	}
	for _, item := range res {
		resp.AppList = append(resp.AppList, &app.AppInfo{
			Id:           int64(item.ID),
			AppName:      item.Name,
			InitPrompt:   item.InitPrompt,
			Cover:        item.Cover,
			DeployKey:    item.Deploykey,
			DeployedTime: item.DeployedTime,
			Priority:     int64(item.Priority),
			UserId:       int64(item.UserId),
			UpdateTime:   item.UpdatedAt.Format("2006-01-02 15:04:05"),
			CreateTime:   item.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return resp, nil
}
