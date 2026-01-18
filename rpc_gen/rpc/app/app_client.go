package app

import (
	"context"
	app "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"

	"github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app/appservice"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/callopt"
)

type RPCClient interface {
	KitexClient() appservice.Client
	Service() string
	AddApp(ctx context.Context, Req *app.AddAppReq, callOptions ...callopt.Option) (r *app.AddAppResp, err error)
	DeleteApp(ctx context.Context, Req *app.DeleteAppReq, callOptions ...callopt.Option) (r *app.DeleteAppResp, err error)
	UpdateApp(ctx context.Context, Req *app.UpdateAppReq, callOptions ...callopt.Option) (r *app.UpdateAppResp, err error)
	GetApp(ctx context.Context, Req *app.GetAppReq, callOptions ...callopt.Option) (r *app.GetAppResp, err error)
	ListApp(ctx context.Context, Req *app.ListAppReq, callOptions ...callopt.Option) (r *app.ListAppResp, err error)
}

func NewRPCClient(dstService string, opts ...client.Option) (RPCClient, error) {
	kitexClient, err := appservice.NewClient(dstService, opts...)
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
	kitexClient appservice.Client
}

func (c *clientImpl) Service() string {
	return c.service
}

func (c *clientImpl) KitexClient() appservice.Client {
	return c.kitexClient
}

func (c *clientImpl) AddApp(ctx context.Context, Req *app.AddAppReq, callOptions ...callopt.Option) (r *app.AddAppResp, err error) {
	return c.kitexClient.AddApp(ctx, Req, callOptions...)
}

func (c *clientImpl) DeleteApp(ctx context.Context, Req *app.DeleteAppReq, callOptions ...callopt.Option) (r *app.DeleteAppResp, err error) {
	return c.kitexClient.DeleteApp(ctx, Req, callOptions...)
}

func (c *clientImpl) UpdateApp(ctx context.Context, Req *app.UpdateAppReq, callOptions ...callopt.Option) (r *app.UpdateAppResp, err error) {
	return c.kitexClient.UpdateApp(ctx, Req, callOptions...)
}

func (c *clientImpl) GetApp(ctx context.Context, Req *app.GetAppReq, callOptions ...callopt.Option) (r *app.GetAppResp, err error) {
	return c.kitexClient.GetApp(ctx, Req, callOptions...)
}

func (c *clientImpl) ListApp(ctx context.Context, Req *app.ListAppReq, callOptions ...callopt.Option) (r *app.ListAppResp, err error) {
	return c.kitexClient.ListApp(ctx, Req, callOptions...)
}
