package service

import (
	"context"

	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
	"github.com/cloudwego/hertz/pkg/app"
)

type SubmitAIService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewSubmitAIService(Context context.Context, RequestContext *app.RequestContext) *SubmitAIService {
	return &SubmitAIService{RequestContext: RequestContext, Context: Context}
}

func (h *SubmitAIService) Run(req *lapp.AISubmitRequest) (resp *lapp.BaseResponseBoolean, err error) {
	ok, err := submitAITask(h.Context, req.GetAppId(), req.GetMessage())
	if err != nil {
		return &lapp.BaseResponseBoolean{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	return &lapp.BaseResponseBoolean{
		Code:    0,
		Data:    ok,
		Message: "success",
	}, nil
}
