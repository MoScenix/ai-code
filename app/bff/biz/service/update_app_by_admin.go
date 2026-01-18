package service

import (
	"context"

	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
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
	//defer func() {
	// hlog.CtxInfof(h.Context, "req = %+v", req)
	// hlog.CtxInfof(h.Context, "resp = %+v", resp)
	//}()
	// todo edit your code
	return
}
