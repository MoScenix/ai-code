package ai

import (
	"context"
	ai "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/ai"
	"github.com/cloudwego/kitex/client/callopt"
	"github.com/cloudwego/kitex/pkg/klog"
)

func Chat(ctx context.Context, req *ai.AiReq, callOptions ...callopt.Option) (resp *ai.AiResp, err error) {
	resp, err = defaultClient.Chat(ctx, req, callOptions...)
	if err != nil {
		klog.CtxErrorf(ctx, "Chat call failed,err =%+v", err)
		return nil, err
	}
	return resp, nil
}
