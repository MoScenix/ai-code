package service

import (
	"context"

	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
	"github.com/cloudwego/hertz/pkg/app"
)

type DeleteAppService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewDeleteAppService(Context context.Context, RequestContext *app.RequestContext) *DeleteAppService {
	return &DeleteAppService{RequestContext: RequestContext, Context: Context}
}

func (h *DeleteAppService) Run(req *lapp.DeleteRequest) (resp *lapp.BaseResponseBoolean, err error) {
	//defer func() {
	// hlog.CtxInfof(h.Context, "req = %+v", req)
	// hlog.CtxInfof(h.Context, "resp = %+v", resp)
	//}()
	// todo edit your code
	return
}
