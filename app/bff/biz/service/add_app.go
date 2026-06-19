package service

import (
	"context"

	"github.com/MoScenix/ai-code/app/bff/biz/utils"
	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
	"github.com/MoScenix/ai-code/app/bff/infra/rpc"
	rpcapp "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/kitex/pkg/klog"
)

type AddAppService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewAddAppService(Context context.Context, RequestContext *app.RequestContext) *AddAppService {
	return &AddAppService{RequestContext: RequestContext, Context: Context}
}

func (h *AddAppService) Run(req *lapp.AppAddRequest) (resp *lapp.BaseResponseLong, err error) {
	userID, _ := h.Context.Value(utils.UserIdKey).(float64)
	ctx := utils.WithIdentityMeta(h.Context)
	res, err := rpc.AppClient.AddApp(ctx, &rpcapp.AddAppReq{
		InitPrompt: req.InitPrompt,
		UserId:     int64(userID),
	})
	if err != nil {
		klog.CtxErrorf(ctx, "create app failed: user_id=%.0f err=%v", userID, err)
		return &lapp.BaseResponseLong{
			Code:    1,
			Message: err.Error(),
		}, err
	}
	klog.CtxInfof(ctx, "app created: app_id=%d user_id=%.0f", res.GetId(), userID)
	return &lapp.BaseResponseLong{
		Code:    0,
		Message: "success",
		Data:    res.Id,
	}, nil
}
