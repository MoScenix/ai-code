package service

import (
	"context"
	"testing"
	app "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
)

func TestListApp_Run(t *testing.T) {
	ctx := context.Background()
	s := NewListAppService(ctx)
	// init req and assert value

	req := &app.ListAppReq{}
	resp, err := s.Run(req)
	t.Logf("err: %v", err)
	t.Logf("resp: %v", resp)

	// todo: edit your unit test

}
