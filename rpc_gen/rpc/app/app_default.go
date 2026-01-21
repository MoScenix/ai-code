package app

import (
	"context"
	app "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
	"github.com/cloudwego/kitex/client/callopt"
	"github.com/cloudwego/kitex/pkg/klog"
)

func AddApp(ctx context.Context, req *app.AddAppReq, callOptions ...callopt.Option) (resp *app.AddAppResp, err error) {
	resp, err = defaultClient.AddApp(ctx, req, callOptions...)
	if err != nil {
		klog.CtxErrorf(ctx, "AddApp call failed,err =%+v", err)
		return nil, err
	}
	return resp, nil
}

func DeleteApp(ctx context.Context, req *app.DeleteAppReq, callOptions ...callopt.Option) (resp *app.DeleteAppResp, err error) {
	resp, err = defaultClient.DeleteApp(ctx, req, callOptions...)
	if err != nil {
		klog.CtxErrorf(ctx, "DeleteApp call failed,err =%+v", err)
		return nil, err
	}
	return resp, nil
}

func UpdateApp(ctx context.Context, req *app.UpdateAppReq, callOptions ...callopt.Option) (resp *app.UpdateAppResp, err error) {
	resp, err = defaultClient.UpdateApp(ctx, req, callOptions...)
	if err != nil {
		klog.CtxErrorf(ctx, "UpdateApp call failed,err =%+v", err)
		return nil, err
	}
	return resp, nil
}

func GetApp(ctx context.Context, req *app.GetAppReq, callOptions ...callopt.Option) (resp *app.GetAppResp, err error) {
	resp, err = defaultClient.GetApp(ctx, req, callOptions...)
	if err != nil {
		klog.CtxErrorf(ctx, "GetApp call failed,err =%+v", err)
		return nil, err
	}
	return resp, nil
}

func ListApp(ctx context.Context, req *app.ListAppReq, callOptions ...callopt.Option) (resp *app.ListAppResp, err error) {
	resp, err = defaultClient.ListApp(ctx, req, callOptions...)
	if err != nil {
		klog.CtxErrorf(ctx, "ListApp call failed,err =%+v", err)
		return nil, err
	}
	return resp, nil
}

func AddMessage(ctx context.Context, req *app.AddMessageReq, callOptions ...callopt.Option) (resp *app.AddMessageResp, err error) {
	resp, err = defaultClient.AddMessage(ctx, req, callOptions...)
	if err != nil {
		klog.CtxErrorf(ctx, "AddMessage call failed,err =%+v", err)
		return nil, err
	}
	return resp, nil
}

func DeleteMessage(ctx context.Context, req *app.DeleteMessageReq, callOptions ...callopt.Option) (resp *app.DeleteMessageResp, err error) {
	resp, err = defaultClient.DeleteMessage(ctx, req, callOptions...)
	if err != nil {
		klog.CtxErrorf(ctx, "DeleteMessage call failed,err =%+v", err)
		return nil, err
	}
	return resp, nil
}

func ListAppMessage(ctx context.Context, req *app.ListAppMessageReq, callOptions ...callopt.Option) (resp *app.ListAppMessageResp, err error) {
	resp, err = defaultClient.ListAppMessage(ctx, req, callOptions...)
	if err != nil {
		klog.CtxErrorf(ctx, "ListAppMessage call failed,err =%+v", err)
		return nil, err
	}
	return resp, nil
}
