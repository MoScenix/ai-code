package document

import (
	"context"
	document "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/document"
	"github.com/cloudwego/kitex/client/callopt"
	"github.com/cloudwego/kitex/pkg/klog"
)

func ParsePDFToText(ctx context.Context, req *document.ParsePDFToTextReq, callOptions ...callopt.Option) (resp *document.ParsePDFToTextResp, err error) {
	resp, err = defaultClient.ParsePDFToText(ctx, req, callOptions...)
	if err != nil {
		klog.CtxErrorf(ctx, "ParsePDFToText call failed,err =%+v", err)
		return nil, err
	}
	return resp, nil
}

func IndexTextFile(ctx context.Context, req *document.IndexTextFileReq, callOptions ...callopt.Option) (resp *document.IndexTextFileResp, err error) {
	resp, err = defaultClient.IndexTextFile(ctx, req, callOptions...)
	if err != nil {
		klog.CtxErrorf(ctx, "IndexTextFile call failed,err =%+v", err)
		return nil, err
	}
	return resp, nil
}

func SearchFile(ctx context.Context, req *document.SearchFileReq, callOptions ...callopt.Option) (resp *document.SearchFileResp, err error) {
	resp, err = defaultClient.SearchFile(ctx, req, callOptions...)
	if err != nil {
		klog.CtxErrorf(ctx, "SearchFile call failed,err =%+v", err)
		return nil, err
	}
	return resp, nil
}

func DeleteProjectFileData(ctx context.Context, req *document.DeleteProjectFileDataReq, callOptions ...callopt.Option) (resp *document.DeleteProjectFileDataResp, err error) {
	resp, err = defaultClient.DeleteProjectFileData(ctx, req, callOptions...)
	if err != nil {
		klog.CtxErrorf(ctx, "DeleteProjectFileData call failed,err =%+v", err)
		return nil, err
	}
	return resp, nil
}
