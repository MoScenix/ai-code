package service

import (
	"context"

	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
	"github.com/cloudwego/hertz/pkg/app"
)

type ListAppVOByPageByAdminService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewListAppVOByPageByAdminService(Context context.Context, RequestContext *app.RequestContext) *ListAppVOByPageByAdminService {
	return &ListAppVOByPageByAdminService{RequestContext: RequestContext, Context: Context}
}

func (h *ListAppVOByPageByAdminService) Run(req *lapp.AppQueryRequest) (resp *lapp.BaseResponsePageAppVO, err error) {
	//defer func() {
	// hlog.CtxInfof(h.Context, "req = %+v", req)
	// hlog.CtxInfof(h.Context, "resp = %+v", resp)
	//}()
	// todo edit your code
	return
}
