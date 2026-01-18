package service

import (
	"context"
	app "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
)

type UpdateAppService struct {
	ctx context.Context
} // NewUpdateAppService new UpdateAppService
func NewUpdateAppService(ctx context.Context) *UpdateAppService {
	return &UpdateAppService{ctx: ctx}
}

// Run create note info
func (s *UpdateAppService) Run(req *app.UpdateAppReq) (resp *app.UpdateAppResp, err error) {
	// Finish your business logic.

	return
}
