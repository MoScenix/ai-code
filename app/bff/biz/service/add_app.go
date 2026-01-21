package service

import (
	"context"

	"github.com/MoScenix/ai-code/app/bff/biz/utils"
	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
	"github.com/MoScenix/ai-code/app/bff/infra/rpc"
	rpcapp "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
	"github.com/cloudwego/hertz/pkg/app"
)

type AddAppService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewAddAppService(Context context.Context, RequestContext *app.RequestContext) *AddAppService {
	return &AddAppService{RequestContext: RequestContext, Context: Context}
}

func (h *AddAppService) Run(req *lapp.AppAddRequest) (resp *lapp.BaseResponseLong, err error) {
	res, err := rpc.AppClient.AddApp(h.Context, &rpcapp.AddAppReq{
		InitPrompt: req.InitPrompt,
		UserId:     int64(h.Context.Value(utils.UserIdKey).(float64)),
	})
	if err != nil {
		return &lapp.BaseResponseLong{
			Code:    1,
			Message: err.Error(),
		}, err
	}
	return &lapp.BaseResponseLong{
		Code:    0,
		Message: "success",
		Data:    res.Id,
	}, nil
}
