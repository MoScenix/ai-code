package user

import (
	"context"
	user "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/user"

	"github.com/MoScenix/ai-code/rpc_gen/kitex_gen/user/userservice"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/callopt"
)

type RPCClient interface {
	KitexClient() userservice.Client
	Service() string
	Login(ctx context.Context, Req *user.LoginReq, callOptions ...callopt.Option) (r *user.LoginResp, err error)
	Register(ctx context.Context, Req *user.RegisterReq, callOptions ...callopt.Option) (r *user.RegisterResp, err error)
	Update(ctx context.Context, Req *user.UpdateReq, callOptions ...callopt.Option) (r *user.UpdateResp, err error)
	GetUser(ctx context.Context, Req *user.GetUserReq, callOptions ...callopt.Option) (r *user.GetUserResp, err error)
	DeleteUser(ctx context.Context, Req *user.DeleteUserReq, callOptions ...callopt.Option) (r *user.DeleteUserResp, err error)
	AddUser(ctx context.Context, Req *user.AddUserReq, callOptions ...callopt.Option) (r *user.AddUserResp, err error)
	ListUser(ctx context.Context, Req *user.ListUserReq, callOptions ...callopt.Option) (r *user.ListUserResp, err error)
}

func NewRPCClient(dstService string, opts ...client.Option) (RPCClient, error) {
	kitexClient, err := userservice.NewClient(dstService, opts...)
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
	kitexClient userservice.Client
}

func (c *clientImpl) Service() string {
	return c.service
}

func (c *clientImpl) KitexClient() userservice.Client {
	return c.kitexClient
}

func (c *clientImpl) Login(ctx context.Context, Req *user.LoginReq, callOptions ...callopt.Option) (r *user.LoginResp, err error) {
	return c.kitexClient.Login(ctx, Req, callOptions...)
}

func (c *clientImpl) Register(ctx context.Context, Req *user.RegisterReq, callOptions ...callopt.Option) (r *user.RegisterResp, err error) {
	return c.kitexClient.Register(ctx, Req, callOptions...)
}

func (c *clientImpl) Update(ctx context.Context, Req *user.UpdateReq, callOptions ...callopt.Option) (r *user.UpdateResp, err error) {
	return c.kitexClient.Update(ctx, Req, callOptions...)
}

func (c *clientImpl) GetUser(ctx context.Context, Req *user.GetUserReq, callOptions ...callopt.Option) (r *user.GetUserResp, err error) {
	return c.kitexClient.GetUser(ctx, Req, callOptions...)
}

func (c *clientImpl) DeleteUser(ctx context.Context, Req *user.DeleteUserReq, callOptions ...callopt.Option) (r *user.DeleteUserResp, err error) {
	return c.kitexClient.DeleteUser(ctx, Req, callOptions...)
}

func (c *clientImpl) AddUser(ctx context.Context, Req *user.AddUserReq, callOptions ...callopt.Option) (r *user.AddUserResp, err error) {
	return c.kitexClient.AddUser(ctx, Req, callOptions...)
}

func (c *clientImpl) ListUser(ctx context.Context, Req *user.ListUserReq, callOptions ...callopt.Option) (r *user.ListUserResp, err error) {
	return c.kitexClient.ListUser(ctx, Req, callOptions...)
}
