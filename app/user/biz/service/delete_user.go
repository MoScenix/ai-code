package service

import (
	"context"

	"github.com/MoScenix/ai-code/app/user/biz/dal/mysql"
	"github.com/MoScenix/ai-code/app/user/biz/model"
	user "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/user"
)

type DeleteUserService struct {
	ctx context.Context
} // NewDeleteUserService new DeleteUserService
func NewDeleteUserService(ctx context.Context) *DeleteUserService {
	return &DeleteUserService{ctx: ctx}
}

// Run create note info
func (s *DeleteUserService) Run(req *user.DeleteUserReq) (resp *user.DeleteUserResp, err error) {
	// Finish your business logic.
	q := model.NewUserQuery(s.ctx, mysql.DB)
	if err := q.DeleteUser(int(req.UserId)); err != nil {
		return nil, err
	}
	return &user.DeleteUserResp{}, nil
}
