package service

import (
	"context"
	app "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
)

type ListAppService struct {
	ctx context.Context
} // NewListAppService new ListAppService
func NewListAppService(ctx context.Context) *ListAppService {
	return &ListAppService{ctx: ctx}
}

// Run create note info
func (s *ListAppService) Run(req *app.ListAppReq) (resp *app.ListAppResp, err error) {
	// Finish your business logic.

	return
}
