package service

import (
	"context"

	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
	"github.com/cloudwego/hertz/pkg/app"
)

type DownloadAppCodeService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewDownloadAppCodeService(Context context.Context, RequestContext *app.RequestContext) *DownloadAppCodeService {
	return &DownloadAppCodeService{RequestContext: RequestContext, Context: Context}
}

func (h *DownloadAppCodeService) Run(req *lapp.DownloadAppCodeRequest) (resp *lapp.BaseResponseBytes, err error) {
	//defer func() {
	// hlog.CtxInfof(h.Context, "req = %+v", req)
	// hlog.CtxInfof(h.Context, "resp = %+v", resp)
	//}()
	// todo edit your code
	return
}
