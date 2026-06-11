package service

import (
	"context"

	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
	"github.com/cloudwego/hertz/pkg/app"
)

type GetAIStateService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewGetAIStateService(Context context.Context, RequestContext *app.RequestContext) *GetAIStateService {
	return &GetAIStateService{RequestContext: RequestContext, Context: Context}
}

func (h *GetAIStateService) Run(req *lapp.AIStateRequest) (resp *lapp.BaseResponseAIState, err error) {
	state, exists, err := loadAIState(h.Context, req.GetAppId())
	if err != nil {
		return &lapp.BaseResponseAIState{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	return &lapp.BaseResponseAIState{
		Code:    0,
		Data:    toAIState(exists, state),
		Message: "success",
	}, nil
}
