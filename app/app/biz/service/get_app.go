package service

import (
	"context"
	app "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
)

type GetAppService struct {
	ctx context.Context
} // NewGetAppService new GetAppService
func NewGetAppService(ctx context.Context) *GetAppService {
	return &GetAppService{ctx: ctx}
}

// Run create note info
func (s *GetAppService) Run(req *app.GetAppReq) (resp *app.GetAppResp, err error) {
	// Finish your business logic.

	return
}
