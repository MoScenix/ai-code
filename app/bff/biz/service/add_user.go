package service

import (
	"context"

	user "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/user"
	"github.com/MoScenix/ai-code/app/bff/infra/rpc"
	rpcuser "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/user"
	"github.com/cloudwego/hertz/pkg/app"
)

type AddUserService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewAddUserService(Context context.Context, RequestContext *app.RequestContext) *AddUserService {
	return &AddUserService{RequestContext: RequestContext, Context: Context}
}

func (h *AddUserService) Run(req *user.UserAddRequest) (resp *user.BaseResponseLong, err error) {
	//defer func() {
	// hlog.CtxInfof(h.Context, "req = %+v", req)
	// hlog.CtxInfof(h.Context, "resp = %+v", resp)
	//}()
	// todo edit your code
	_, err = rpc.UserClient.AddUser(h.Context, &rpcuser.AddUserReq{
		UserAccount:  req.UserAccount,
		UserAvatar:   req.UserAvatar,
		UserName:     req.UserName,
		UserProfile:  req.UserProfile,
		UserRole:     req.UserRole,
		UserPassword: req.UserName,
	})
	if err != nil {
		return &user.BaseResponseLong{
			Code:    1,
			Message: err.Error(),
		}, err
	}
	return &user.BaseResponseLong{
		Code:    0,
		Message: "success",
	}, nil
}
