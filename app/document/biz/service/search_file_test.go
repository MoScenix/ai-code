package service

import (
	"context"
	"testing"
	document "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/document"
)

func TestSearchFile_Run(t *testing.T) {
	ctx := context.Background()
	s := NewSearchFileService(ctx)
	// init req and assert value

	req := &document.SearchFileReq{}
	resp, err := s.Run(req)
	t.Logf("err: %v", err)
	t.Logf("resp: %v", resp)

	// todo: edit your unit test

}
