package service

import (
	"context"
	"testing"
	app "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
)

func TestAddApp_Run(t *testing.T) {
	ctx := context.Background()
	s := NewAddAppService(ctx)
	// init req and assert value

	req := &app.AddAppReq{}
	resp, err := s.Run(req)
	t.Logf("err: %v", err)
	t.Logf("resp: %v", resp)

	// todo: edit your unit test

}
