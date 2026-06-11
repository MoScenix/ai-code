package service

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	ai "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/ai"
	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2/metadata"
)

type fakeChatStream struct {
	ctx context.Context
}

// Header implements [ai.AiService_ChatServer].
func (f *fakeChatStream) Header() (metadata.MD, error) {
	panic("unimplemented")
}

// RecvMsg implements [ai.AiService_ChatServer].
func (f *fakeChatStream) RecvMsg(m interface{}) error {
	panic("unimplemented")
}

// SendHeader implements [ai.AiService_ChatServer].
func (f *fakeChatStream) SendHeader(metadata.MD) error {
	panic("unimplemented")
}

// SendMsg implements [ai.AiService_ChatServer].
func (f *fakeChatStream) SendMsg(m interface{}) error {
	panic("unimplemented")
}

// SetHeader implements [ai.AiService_ChatServer].
func (f *fakeChatStream) SetHeader(metadata.MD) error {
	panic("unimplemented")
}

// SetTrailer implements [ai.AiService_ChatServer].
func (f *fakeChatStream) SetTrailer(metadata.MD) {
	panic("unimplemented")
}

// Trailer implements [ai.AiService_ChatServer].
func (f *fakeChatStream) Trailer() metadata.MD {
	panic("unimplemented")
}

func newFakeChatStream() *fakeChatStream {
	return &fakeChatStream{ctx: context.Background()}
}

func (f *fakeChatStream) Context() context.Context { return f.ctx }

func (f *fakeChatStream) Send(resp *ai.AiResp) error {
	return nil
}
func (f *fakeChatStream) Close() error {
	return nil
}
func (f *fakeChatStream) Recv() (*ai.AiReq, error) {
	return nil, io.EOF
}

func TestChat_Run(t *testing.T) {
	shareDir := filepath.Join(t.TempDir(), "project")
	confPath := filepath.Join(t.TempDir(), "filestore.yaml")
	content := []byte("ShareDir:\n  share_dir: " + shareDir + "\ncache:\n  cache_dir: " + filepath.Join(shareDir, "cache") + "\n  ttl_seconds: 7200\n  need_flush: true\n")
	if err := os.WriteFile(confPath, content, 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("FILESTORE_CONF_PATH", confPath)

	req := &ai.AiReq{
		ProjectId: "demo",
	}
	var a = fakeChatStream{}
	err := NewChatService(context.Background()).Run(req.GetProjectId(), &a)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}
