package service

import (
	"context"
	"testing"
	app "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
)

func TestDeleteApp_Run(t *testing.T) {
	ctx := context.Background()
	s := NewDeleteAppService(ctx)
	// init req and assert value

	req := &app.DeleteAppReq{}
	resp, err := s.Run(req)
	t.Logf("err: %v", err)
	t.Logf("resp: %v", resp)

	// todo: edit your unit test

}
