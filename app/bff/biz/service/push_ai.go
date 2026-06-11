package service

import (
	"context"

	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
	"github.com/cloudwego/hertz/pkg/app"
)

type PushAIService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewPushAIService(Context context.Context, RequestContext *app.RequestContext) *PushAIService {
	return &PushAIService{RequestContext: RequestContext, Context: Context}
}

func (h *PushAIService) Run(req *lapp.AIControlRequest) (resp *lapp.BaseResponseString, err error) {
	id, err := pushAIEvent(h.Context, req.GetAppId(), req.GetContent())
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
