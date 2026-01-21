package service

import (
	"context"

	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
	"github.com/MoScenix/ai-code/app/bff/infra/rpc"
	rpcapp "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
	rpcuser "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/user"
	"github.com/cloudwego/hertz/pkg/app"
)

type ListGoodAppVOByPageService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewListGoodAppVOByPageService(Context context.Context, RequestContext *app.RequestContext) *ListGoodAppVOByPageService {
	return &ListGoodAppVOByPageService{RequestContext: RequestContext, Context: Context}
}

func (h *ListGoodAppVOByPageService) Run(req *lapp.AppQueryRequest) (resp *lapp.BaseResponsePageAppVO, err error) {
	res, err := rpc.AppClient.ListApp(h.Context, &rpcapp.ListAppReq{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		UserId:   req.UserId,
		AppName:  req.AppName,
	})
	if err != nil {
		return &lapp.BaseResponsePageAppVO{
			Code:    1,
			Message: err.Error(),
		}, err
	}
	resp = &lapp.BaseResponsePageAppVO{
		Code:    0,
		Message: "success",
		Data: &lapp.PageAppVO{
			TotalPage:  res.Total,
			PageSize:   req.PageSize,
			PageNumber: req.PageNum,
			Records:    []*lapp.AppVO{},
		},
	}
	q := rpc.UserClient
	for _, app := range res.AppList {
		r, err := q.GetUser(h.Context, &rpcuser.GetUserReq{
			Id: app.UserId,
		})
		if err != nil {
			return &lapp.BaseResponsePageAppVO{
				Code:    1,
				Message: err.Error(),
			}, err
		}
		resp.Data.Records = append(resp.Data.Records, &lapp.AppVO{
			Id:           app.Id,
			AppName:      app.AppName,
			CreateTime:   app.CreateTime,
			UpdateTime:   app.UpdateTime,
			UserId:       app.UserId,
			DeployKey:    app.DeployKey,
			DeployedTime: app.DeployedTime,
			Priority:     app.Priority,
			Cover:        app.Cover,
			InitPrompt:   app.InitPrompt,
			User: &lapp.UserVO{
				Id:          r.Id,
				UserName:    r.UserName,
				UserAccount: r.UserAccount,
				UserAvatar:  r.UserAvatar,
				UserProfile: r.UserProfile,
				UserRole:    r.UserRole,
				CreateTime:  r.CreateTime,
			},
		})
	}
	return resp, nil
}
