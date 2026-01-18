package service

import (
	"context"

	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
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
	//defer func() {
	// hlog.CtxInfof(h.Context, "req = %+v", req)
	// hlog.CtxInfof(h.Context, "resp = %+v", resp)
	//}()
	// todo edit your code
	return
}
