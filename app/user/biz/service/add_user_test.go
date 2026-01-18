package service

import (
	"context"
	"testing"
	user "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/user"
)

func TestAddUser_Run(t *testing.T) {
	ctx := context.Background()
	s := NewAddUserService(ctx)
	// init req and assert value

	req := &user.AddUserReq{}
	resp, err := s.Run(req)
	t.Logf("err: %v", err)
	t.Logf("resp: %v", resp)

	// todo: edit your unit test

}
