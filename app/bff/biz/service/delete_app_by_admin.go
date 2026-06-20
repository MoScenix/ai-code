package service

import (
	"context"

	"github.com/MoScenix/ai-code/app/bff/biz/utils"
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
	ctx := utils.WithIdentityMeta(h.Context)
	res, err := rpc.AppClient.DeleteApp(ctx, &rpcapp.DeleteAppReq{
		Id: req.Id,
	})
	if err != nil {
		return &lapp.BaseResponseBoolean{
			Code:    1,
			Message: err.Error(),
		}, err
	}
	if res.Success {
		if err := deleteProjectFileData(ctx, req.Id); err != nil {
			return &lapp.BaseResponseBoolean{
				Code:    1,
				Message: err.Error(),
			}, err
		}
	}
	return &lapp.BaseResponseBoolean{
		Code:    0,
		Message: "success",
		Data:    res.Success,
	}, nil
}
