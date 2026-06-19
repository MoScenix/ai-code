package service

import (
	"context"
	"testing"
	document "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/document"
)

func TestDeleteProjectFileData_Run(t *testing.T) {
	ctx := context.Background()
	s := NewDeleteProjectFileDataService(ctx)
	// init req and assert value

	req := &document.DeleteProjectFileDataReq{}
	resp, err := s.Run(req)
	t.Logf("err: %v", err)
	t.Logf("resp: %v", resp)

	// todo: edit your unit test

}
