package service

import (
	"context"

	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
	"github.com/cloudwego/hertz/pkg/app"
)

type GetAppVOByIdService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewGetAppVOByIdService(Context context.Context, RequestContext *app.RequestContext) *GetAppVOByIdService {
	return &GetAppVOByIdService{RequestContext: RequestContext, Context: Context}
}

func (h *GetAppVOByIdService) Run(req *lapp.GetAppVOByIdRequest) (resp *lapp.BaseResponseAppVO, err error) {
	//defer func() {
	// hlog.CtxInfof(h.Context, "req = %+v", req)
	// hlog.CtxInfof(h.Context, "resp = %+v", resp)
	//}()
	// todo edit your code
	return
}
