package service

import (
	"context"

	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
	"github.com/MoScenix/ai-code/app/bff/infra/rpc"
	rpcapp "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
	"github.com/cloudwego/hertz/pkg/app"
)

type UpdateAppByAdminService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewUpdateAppByAdminService(Context context.Context, RequestContext *app.RequestContext) *UpdateAppByAdminService {
	return &UpdateAppByAdminService{RequestContext: RequestContext, Context: Context}
}

func (h *UpdateAppByAdminService) Run(req *lapp.AppAdminUpdateRequest) (resp *lapp.BaseResponseBoolean, err error) {
	res, err := rpc.AppClient.UpdateApp(h.Context, &rpcapp.UpdateAppReq{
		Id:       req.Id,
		AppName:  req.AppName,
		Cover:    req.Cover,
		Priority: req.Priority,
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
	return
}
