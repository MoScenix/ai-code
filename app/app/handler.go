package main

import (
	"context"
	app "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
	"github.com/MoScenix/ai-code/app/app/biz/service"
)

// AppServiceImpl implements the last service interface defined in the IDL.
type AppServiceImpl struct{}

// AddApp implements the AppServiceImpl interface.
func (s *AppServiceImpl) AddApp(ctx context.Context, req *app.AddAppReq) (resp *app.AddAppResp, err error) {
	resp, err = service.NewAddAppService(ctx).Run(req)

	return resp, err
}

// DeleteApp implements the AppServiceImpl interface.
func (s *AppServiceImpl) DeleteApp(ctx context.Context, req *app.DeleteAppReq) (resp *app.DeleteAppResp, err error) {
	resp, err = service.NewDeleteAppService(ctx).Run(req)

	return resp, err
}

// UpdateApp implements the AppServiceImpl interface.
func (s *AppServiceImpl) UpdateApp(ctx context.Context, req *app.UpdateAppReq) (resp *app.UpdateAppResp, err error) {
	resp, err = service.NewUpdateAppService(ctx).Run(req)

	return resp, err
}

// GetApp implements the AppServiceImpl interface.
func (s *AppServiceImpl) GetApp(ctx context.Context, req *app.GetAppReq) (resp *app.GetAppResp, err error) {
	resp, err = service.NewGetAppService(ctx).Run(req)

	return resp, err
}

// ListApp implements the AppServiceImpl interface.
func (s *AppServiceImpl) ListApp(ctx context.Context, req *app.ListAppReq) (resp *app.ListAppResp, err error) {
	resp, err = service.NewListAppService(ctx).Run(req)

	return resp, err
}

// AddMessage implements the AppServiceImpl interface.
func (s *AppServiceImpl) AddMessage(ctx context.Context, req *app.AddMessageReq) (resp *app.AddMessageResp, err error) {
	resp, err = service.NewAddMessageService(ctx).Run(req)

	return resp, err
}

// DeleteMessage implements the AppServiceImpl interface.
func (s *AppServiceImpl) DeleteMessage(ctx context.Context, req *app.DeleteMessageReq) (resp *app.DeleteMessageResp, err error) {
	resp, err = service.NewDeleteMessageService(ctx).Run(req)

	return resp, err
}

// ListAppMessage implements the AppServiceImpl interface.
func (s *AppServiceImpl) ListAppMessage(ctx context.Context, req *app.ListAppMessageReq) (resp *app.ListAppMessageResp, err error) {
	resp, err = service.NewListAppMessageService(ctx).Run(req)

	return resp, err
}
