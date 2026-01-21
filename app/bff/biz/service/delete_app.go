package service

import (
	"context"

	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
	"github.com/MoScenix/ai-code/app/bff/infra/rpc"
	rpcapp "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
	"github.com/cloudwego/hertz/pkg/app"
)

type DeleteAppService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewDeleteAppService(Context context.Context, RequestContext *app.RequestContext) *DeleteAppService {
	return &DeleteAppService{RequestContext: RequestContext, Context: Context}
}

func (h *DeleteAppService) Run(req *lapp.DeleteRequest) (resp *lapp.BaseResponseBoolean, err error) {
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
