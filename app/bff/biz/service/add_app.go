package service

import (
	"context"

	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
	"github.com/cloudwego/hertz/pkg/app"
)

type AddAppService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewAddAppService(Context context.Context, RequestContext *app.RequestContext) *AddAppService {
	return &AddAppService{RequestContext: RequestContext, Context: Context}
}

func (h *AddAppService) Run(req *lapp.AppAddRequest) (resp *lapp.BaseResponseLong, err error) {

	return
}
