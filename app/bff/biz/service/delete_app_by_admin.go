package service

import (
	"context"

	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
	"github.com/cloudwego/hertz/pkg/app"
)

type DeleteAppByAdminService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewDeleteAppByAdminService(Context context.Context, RequestContext *app.RequestContext) *DeleteAppByAdminService {
	return &DeleteAppByAdminService{RequestContext: RequestContext, Context: Context}
}

func (h *DeleteAppByAdminService) Run(req *lapp.DeleteRequest) (resp *lapp.BaseResponseBoolean, err error) {
	//defer func() {
	// hlog.CtxInfof(h.Context, "req = %+v", req)
	// hlog.CtxInfof(h.Context, "resp = %+v", resp)
	//}()
	// todo edit your code
	return
}
