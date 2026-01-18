package service

import (
	"context"

	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
	"github.com/cloudwego/hertz/pkg/app"
)

type DeployAppService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewDeployAppService(Context context.Context, RequestContext *app.RequestContext) *DeployAppService {
	return &DeployAppService{RequestContext: RequestContext, Context: Context}
}

func (h *DeployAppService) Run(req *lapp.AppDeployRequest) (resp *lapp.BaseResponseString, err error) {
	//defer func() {
	// hlog.CtxInfof(h.Context, "req = %+v", req)
	// hlog.CtxInfof(h.Context, "resp = %+v", resp)
	//}()
	// todo edit your code
	return
}
