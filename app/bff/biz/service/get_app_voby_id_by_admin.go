package service

import (
	"context"

	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
	"github.com/MoScenix/ai-code/app/bff/infra/rpc"
	rpcapp "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
	rpcuser "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/user"
	"github.com/cloudwego/hertz/pkg/app"
)

type GetAppVOByIdByAdminService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewGetAppVOByIdByAdminService(Context context.Context, RequestContext *app.RequestContext) *GetAppVOByIdByAdminService {
	return &GetAppVOByIdByAdminService{RequestContext: RequestContext, Context: Context}
}

func (h *GetAppVOByIdByAdminService) Run(req *lapp.GetAppVOByIdByAdminRequest) (resp *lapp.BaseResponseAppVO, err error) {
	res, err := rpc.AppClient.GetApp(h.Context, &rpcapp.GetAppReq{
		Id: req.Id,
	})
	if err != nil {
		return &lapp.BaseResponseAppVO{
			Code:    1,
			Message: err.Error(),
		}, err
	}
	r, err := rpc.UserClient.GetUser(h.Context, &rpcuser.GetUserReq{
		Id: res.App.UserId,
	})
	if err != nil {
		return &lapp.BaseResponseAppVO{
			Code:    1,
			Message: err.Error(),
		}, err
	}
	return &lapp.BaseResponseAppVO{
		Code:    0,
		Message: "success",
		Data: &lapp.AppVO{
			Id:           res.App.Id,
			AppName:      res.App.AppName,
			CreateTime:   res.App.CreateTime,
			UpdateTime:   res.App.UpdateTime,
			UserId:       res.App.UserId,
			DeployKey:    res.App.DeployKey,
			DeployedTime: res.App.DeployedTime,
			Priority:     res.App.Priority,
			Cover:        res.App.Cover,
			InitPrompt:   res.App.InitPrompt,
			User: &lapp.UserVO{
				Id:          r.Id,
				UserName:    r.UserName,
				UserAccount: r.UserAccount,
				UserAvatar:  r.UserAvatar,
				UserProfile: r.UserProfile,
				UserRole:    r.UserRole,
				CreateTime:  r.CreateTime,
			},
		},
	}, nil
}
