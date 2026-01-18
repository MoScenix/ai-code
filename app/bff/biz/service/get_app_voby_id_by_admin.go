package service

import (
	"context"

	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
	"github.com/cloudwego/hertz/pkg/app"
)

type GetAppVOByIdByAdminService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewGetAppVOByIdByAdminService(Context context.Context, RequestContext *app.RequestContext) *GetAppVOByIdByAdminService {
	return &GetAppVOByIdByAdminService{RequestContext: RequestContext, Context: Context}
}

func (h *GetAppVOByIdByAdminService) Run(req *lapp.GetAppVOByIdByAdminRequest) (resp *lapp.BaseResponseAppVO, err error) {
	//defer func() {
	// hlog.CtxInfof(h.Context, "req = %+v", req)
	// hlog.CtxInfof(h.Context, "resp = %+v", resp)
	//}()
	// todo edit your code
	return
}
