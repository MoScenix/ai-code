package service

import (
	"context"

	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
	"github.com/cloudwego/hertz/pkg/app"
)

type ListGoodAppVOByPageService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewListGoodAppVOByPageService(Context context.Context, RequestContext *app.RequestContext) *ListGoodAppVOByPageService {
	return &ListGoodAppVOByPageService{RequestContext: RequestContext, Context: Context}
}

func (h *ListGoodAppVOByPageService) Run(req *lapp.AppQueryRequest) (resp *lapp.BaseResponsePageAppVO, err error) {
	//defer func() {
	// hlog.CtxInfof(h.Context, "req = %+v", req)
	// hlog.CtxInfof(h.Context, "resp = %+v", resp)
	//}()
	// todo edit your code
	return
}
