package ai

import (
	ai "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/ai"

	"github.com/MoScenix/ai-code/rpc_gen/kitex_gen/ai/aiservice"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/callopt"
)

type RPCClient interface {
	KitexClient() aiservice.Client
	Service() string
	Chat(ctx context.Context, Req *ai.AiReq, callOptions ...callopt.Option) (stream aiservice.AiService_ChatClient, err error)
}

func NewRPCClient(dstService string, opts ...client.Option) (RPCClient, error) {
	kitexClient, err := aiservice.NewClient(dstService, opts...)
	if err != nil {
		return nil, err
	}
	cli := &clientImpl{
		service:     dstService,
		kitexClient: kitexClient,
	}

	return cli, nil
}

type clientImpl struct {
	service     string
	kitexClient aiservice.Client
}

func (c *clientImpl) Service() string {
	return c.service
}

func (c *clientImpl) KitexClient() aiservice.Client {
	return c.kitexClient
}

func (c *clientImpl) Chat(ctx context.Context, Req *ai.AiReq, callOptions ...callopt.Option) (stream aiservice.AiService_ChatClient, err error) {
	return c.kitexClient.Chat(ctx, Req, callOptions...)
}
