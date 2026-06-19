package service

import (
	"context"

	"github.com/MoScenix/ai-code/app/bff/biz/utils"
	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
	"github.com/MoScenix/ai-code/app/bff/infra/rpc"
	rpcapp "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
	"github.com/cloudwego/hertz/pkg/app"
)

type UpdateAppService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewUpdateAppService(Context context.Context, RequestContext *app.RequestContext) *UpdateAppService {
	return &UpdateAppService{RequestContext: RequestContext, Context: Context}
}

func (h *UpdateAppService) Run(req *lapp.AppUpdateRequest) (resp *lapp.BaseResponseBoolean, err error) {
	ctx := utils.WithIdentityMeta(h.Context)
	res, err := rpc.AppClient.UpdateApp(ctx, &rpcapp.UpdateAppReq{
		Id:      req.Id,
		AppName: req.AppName,
	})
	if err != nil {
		return &lapp.BaseResponseBoolean{
			Code:    1,
			Message: err.Error(),
		}, err
	}
	return &lapp.BaseResponseBoolean{
		Code:    0,
		Message: "success",
		Data:    res.Success,
	}, nil
}
