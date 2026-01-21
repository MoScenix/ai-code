package service

import (
	"context"

	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
	"github.com/MoScenix/ai-code/app/bff/infra/rpc"
	rpcapp "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
	"github.com/cloudwego/hertz/pkg/app"
)

type DeleteAppByAdminService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewDeleteAppByAdminService(Context context.Context, RequestContext *app.RequestContext) *DeleteAppByAdminService {
	return &DeleteAppByAdminService{RequestContext: RequestContext, Context: Context}
}

func (h *DeleteAppByAdminService) Run(req *lapp.DeleteRequest) (resp *lapp.BaseResponseBoolean, err error) {
	res, err := rpc.AppClient.DeleteApp(h.Context, &rpcapp.DeleteAppReq{
		Id: req.Id,
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
