package service

import (
	"context"

	"github.com/MoScenix/ai-code/app/user/biz/dal/mysql"
	"github.com/MoScenix/ai-code/app/user/biz/model"
	user "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/user"
	"golang.org/x/crypto/bcrypt"
)

type LoginService struct {
	ctx context.Context
} // NewLoginService new LoginService
func NewLoginService(ctx context.Context) *LoginService {
	return &LoginService{ctx: ctx}
}

// Run create note info
func (s *LoginService) Run(req *user.LoginReq) (resp *user.LoginResp, err error) {
	// Finish your business logic.
	q := model.NewUserQuery(s.ctx, mysql.DB)
	User, err := q.GetUserByAccount(req.UserAccount)
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(User.PasswordHash), []byte(req.UserPassword)); err != nil {
		return nil, err
	}
	return &user.LoginResp{
		UserId:   uint32(User.ID),
		UserRole: User.UserRole,
	}, nil
}
