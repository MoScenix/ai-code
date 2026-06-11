package service

import (
	"context"

	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
	"github.com/cloudwego/hertz/pkg/app"
)

type ListAIEventsService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewListAIEventsService(Context context.Context, RequestContext *app.RequestContext) *ListAIEventsService {
	return &ListAIEventsService{RequestContext: RequestContext, Context: Context}
}

func (h *ListAIEventsService) Run(req *lapp.AIEventsRequest) (resp *lapp.BaseResponseAIEvents, err error) {
	events, err := listAIEvents(h.Context, req.GetAppId(), req.GetLastId(), req.GetBlockMs(), req.GetCount())
	if err != nil {
		return &lapp.BaseResponseAIEvents{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	return &lapp.BaseResponseAIEvents{
		Code:    0,
		Data:    events,
		Message: "success",
	}, nil
}
