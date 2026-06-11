package service

import (
	"context"

	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
	"github.com/cloudwego/hertz/pkg/app"
)

type CancelAIService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewCancelAIService(Context context.Context, RequestContext *app.RequestContext) *CancelAIService {
	return &CancelAIService{RequestContext: RequestContext, Context: Context}
}

func (h *CancelAIService) Run(req *lapp.AIControlRequest) (resp *lapp.BaseResponseString, err error) {
	id, err := cancelAIEvent(h.Context, req.GetAppId(), req.GetReason())
	if err != nil {
		return &lapp.BaseResponseString{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	return &lapp.BaseResponseString{
		Code:    0,
		Data:    id,
		Message: "success",
	}, nil
}
