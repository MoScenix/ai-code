package service

import (
	"context"

	"github.com/MoScenix/ai-code/app/user/biz/dal/mysql"
	"github.com/MoScenix/ai-code/app/user/biz/model"
	user "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/user"
)

type UpdateService struct {
	ctx context.Context
} // NewUpdateService new UpdateService
func NewUpdateService(ctx context.Context) *UpdateService {
	return &UpdateService{ctx: ctx}
}

// Run create note info
func (s *UpdateService) Run(req *user.UpdateReq) (resp *user.UpdateResp, err error) {
	// Finish your business logic.
	err = model.NewUserQuery(s.ctx, mysql.DB).UpdateUser(uint(req.Id), model.User{
		UserAvatar:  req.UserAvatar,
		Name:        req.UserName,
		UserProfile: req.UserProfile,
	})
	if err != nil {
		return nil, err
	}
	return &user.UpdateResp{}, nil
}
