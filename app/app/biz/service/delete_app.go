package service

import (
	"context"

	"github.com/MoScenix/ai-code/app/app/biz/dal/mysql"
	"github.com/MoScenix/ai-code/app/app/biz/model"
	app "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
)

type DeleteAppService struct {
	ctx context.Context
} // NewDeleteAppService new DeleteAppService
func NewDeleteAppService(ctx context.Context) *DeleteAppService {
	return &DeleteAppService{ctx: ctx}
}

// Run create note info
func (s *DeleteAppService) Run(req *app.DeleteAppReq) (resp *app.DeleteAppResp, err error) {
	// Finish your business logic.
	err = model.NewAppQuery(s.ctx, mysql.DB).DeleteApp(uint(req.Id))
	if err != nil {
		return nil, err
	}
	return &app.DeleteAppResp{
		Success: true,
	}, nil
}
