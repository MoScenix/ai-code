package service

import (
	"context"
	"testing"
	document "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/document"
)

func TestParsePDFToText_Run(t *testing.T) {
	ctx := context.Background()
	s := NewParsePDFToTextService(ctx)
	// init req and assert value

	req := &document.ParsePDFToTextReq{}
	resp, err := s.Run(req)
	t.Logf("err: %v", err)
	t.Logf("resp: %v", resp)

	// todo: edit your unit test

}
