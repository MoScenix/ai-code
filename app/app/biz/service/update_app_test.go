package service

import (
	"context"
	"testing"
	app "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
)

func TestUpdateApp_Run(t *testing.T) {
	ctx := context.Background()
	s := NewUpdateAppService(ctx)
	// init req and assert value

	req := &app.UpdateAppReq{}
	resp, err := s.Run(req)
	t.Logf("err: %v", err)
	t.Logf("resp: %v", resp)

	// todo: edit your unit test

}
