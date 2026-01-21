package service

import (
	"context"
	"time"

	"github.com/MoScenix/ai-code/app/app/biz/dal/mysql"
	"github.com/MoScenix/ai-code/app/app/biz/model"
	app "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
)

type ListAppMessageService struct {
	ctx context.Context
} // NewListAppMessageService new ListAppMessageService
func NewListAppMessageService(ctx context.Context) *ListAppMessageService {
	return &ListAppMessageService{ctx: ctx}
}

// Run create note info
func (s *ListAppMessageService) Run(req *app.ListAppMessageReq) (resp *app.ListAppMessageResp, err error) {
	// Finish your business logic.
	q := model.NewMessageQuery(s.ctx, mysql.DB)
	tot, err := q.Count(uint(req.AppId))
	if err != nil {
		return nil, err
	}
	resp = &app.ListAppMessageResp{
		Total: int64(tot),
	}
	t, err := time.Parse("2006-01-02 15:04:05", req.LastCreateTime)
	if err != nil {
		return nil, err
	}
	Appres, err := model.NewAppQuery(s.ctx, mysql.DB).GetAppById(uint(req.AppId))
	if err != nil {
		return nil, err
	}
	res, err := q.ListMessagesByAppId(uint(req.AppId), int(req.PageSize), &t)
	for _, v := range res {
		resp.MessageList = append(resp.MessageList, &app.AppMessage{
			Id:         int64(v.ID),
			AppId:      int64(v.AppId),
			Content:    v.Content,
			Role:       v.Role,
			CreateTime: v.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdateTime: v.CreatedAt.Format("2006-01-02 15:04:05"),
			UserId:     int64(Appres.UserId),
		})
	}
	return resp, nil
}
