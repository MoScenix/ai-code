package service

import (
	"context"

	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
	"github.com/cloudwego/hertz/pkg/app"
)

type AnswerAIService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewAnswerAIService(Context context.Context, RequestContext *app.RequestContext) *AnswerAIService {
	return &AnswerAIService{RequestContext: RequestContext, Context: Context}
}

func (h *AnswerAIService) Run(req *lapp.AIControlRequest) (resp *lapp.BaseResponseBoolean, err error) {
	submitted, err := answerAIQuestion(h.Context, req.GetAppId(), req.GetContent(), req.GetTargetId())
	if err != nil {
		return &lapp.BaseResponseBoolean{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	return &lapp.BaseResponseBoolean{
		Code:    0,
		Data:    submitted,
		Message: "success",
	}, nil
}
