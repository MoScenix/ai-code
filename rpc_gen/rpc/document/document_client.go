package document

import (
	"context"
	document "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/document"

	"github.com/MoScenix/ai-code/rpc_gen/kitex_gen/document/documentservice"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/callopt"
)

type RPCClient interface {
	KitexClient() documentservice.Client
	Service() string
	ParsePDFToText(ctx context.Context, Req *document.ParsePDFToTextReq, callOptions ...callopt.Option) (r *document.ParsePDFToTextResp, err error)
	IndexTextFile(ctx context.Context, Req *document.IndexTextFileReq, callOptions ...callopt.Option) (r *document.IndexTextFileResp, err error)
	SearchFile(ctx context.Context, Req *document.SearchFileReq, callOptions ...callopt.Option) (r *document.SearchFileResp, err error)
	DeleteProjectFileData(ctx context.Context, Req *document.DeleteProjectFileDataReq, callOptions ...callopt.Option) (r *document.DeleteProjectFileDataResp, err error)
}

func NewRPCClient(dstService string, opts ...client.Option) (RPCClient, error) {
	kitexClient, err := documentservice.NewClient(dstService, opts...)
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
	kitexClient documentservice.Client
}

func (c *clientImpl) Service() string {
	return c.service
}

func (c *clientImpl) KitexClient() documentservice.Client {
	return c.kitexClient
}

func (c *clientImpl) ParsePDFToText(ctx context.Context, Req *document.ParsePDFToTextReq, callOptions ...callopt.Option) (r *document.ParsePDFToTextResp, err error) {
	return c.kitexClient.ParsePDFToText(ctx, Req, callOptions...)
}

func (c *clientImpl) IndexTextFile(ctx context.Context, Req *document.IndexTextFileReq, callOptions ...callopt.Option) (r *document.IndexTextFileResp, err error) {
	return c.kitexClient.IndexTextFile(ctx, Req, callOptions...)
}

func (c *clientImpl) SearchFile(ctx context.Context, Req *document.SearchFileReq, callOptions ...callopt.Option) (r *document.SearchFileResp, err error) {
	return c.kitexClient.SearchFile(ctx, Req, callOptions...)
}

func (c *clientImpl) DeleteProjectFileData(ctx context.Context, Req *document.DeleteProjectFileDataReq, callOptions ...callopt.Option) (r *document.DeleteProjectFileDataResp, err error) {
	return c.kitexClient.DeleteProjectFileData(ctx, Req, callOptions...)
}
