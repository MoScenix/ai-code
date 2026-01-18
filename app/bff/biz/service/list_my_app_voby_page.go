package service

import (
	"context"

	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
	"github.com/cloudwego/hertz/pkg/app"
)

type ListMyAppVOByPageService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewListMyAppVOByPageService(Context context.Context, RequestContext *app.RequestContext) *ListMyAppVOByPageService {
	return &ListMyAppVOByPageService{RequestContext: RequestContext, Context: Context}
}

func (h *ListMyAppVOByPageService) Run(req *lapp.AppQueryRequest) (resp *lapp.BaseResponsePageAppVO, err error) {
	//defer func() {
	// hlog.CtxInfof(h.Context, "req = %+v", req)
	// hlog.CtxInfof(h.Context, "resp = %+v", resp)
	//}()
	// todo edit your code
	return
}
