package service

import (
	"context"
	app "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
)

type AddAppService struct {
	ctx context.Context
} // NewAddAppService new AddAppService
func NewAddAppService(ctx context.Context) *AddAppService {
	return &AddAppService{ctx: ctx}
}

// Run create note info
func (s *AddAppService) Run(req *app.AddAppReq) (resp *app.AddAppResp, err error) {
	// Finish your business logic.

	return
}
