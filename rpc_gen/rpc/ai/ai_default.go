package ai

import (
	ai "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/ai"
	"github.com/MoScenix/ai-code/rpc_gen/kitex_gen/ai/aiservice"

	"github.com/cloudwego/kitex/client/callopt"
	"github.com/cloudwego/kitex/pkg/klog"
)

func Chat(ctx context.Context, Req *ai.AiReq, callOptions ...callopt.Option) (stream aiservice.AiService_ChatClient, err error) {
	stream, err = defaultClient.Chat(ctx, Req, callOptions...)
	if err != nil {
		klog.CtxErrorf(ctx, "Chat call failed,err =%+v", err)
		return nil, err
	}
	return stream, nil
}
