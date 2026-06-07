package service

import (
	"context"
	"errors"
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
	return nil, ErrChatRuntimeUnavailable
}

func TestChat_Run(t *testing.T) {
	req := &ai.AiReq{
		ProjectId: "demo",
		History: []*ai.HistoryItem{
			{Role: "user", Question: `测试，写一个upcpc竞赛宣传网页`},
		},
	}
	var a = fakeChatStream{}
	err := NewChatService(context.Background()).Run(req, &a)
	if !errors.Is(err, ErrChatRuntimeUnavailable) {
		t.Fatalf("expected ErrChatRuntimeUnavailable, got %v", err)
	}
}
