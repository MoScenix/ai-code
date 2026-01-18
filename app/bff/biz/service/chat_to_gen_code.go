package service

import (
	"context"

	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
	"github.com/cloudwego/hertz/pkg/app"
)

type ChatToGenCodeService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewChatToGenCodeService(Context context.Context, RequestContext *app.RequestContext) *ChatToGenCodeService {
	return &ChatToGenCodeService{RequestContext: RequestContext, Context: Context}
}

func (h *ChatToGenCodeService) Run(req *lapp.ChatToGenCodeRequest) (resp *lapp.ServerSentEventStringList, err error) {
	//defer func() {
	// hlog.CtxInfof(h.Context, "req = %+v", req)
	// hlog.CtxInfof(h.Context, "resp = %+v", resp)
	//}()
	// todo edit your code
	return
}
