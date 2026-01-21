package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	ai "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/ai"
	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2/metadata"
	"github.com/joho/godotenv"
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
	fmt.Printf("%+v\n", resp)
	return nil
}
func (f *fakeChatStream) Close() error {
	return nil
}
func (f *fakeChatStream) Recv() (*ai.AiReq, error) {
	return nil, fmt.Errorf("Recv not implemented for testing")
}

func TestChat_Run(t *testing.T) {
	// todo: edit your unit test
	_, thisFile, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(thisFile)
	target := filepath.Clean(filepath.Join(baseDir, "../../"))
	godotenv.Load()
	req := &ai.AiReq{
		ProjectId: "7",
		History: []*ai.HistoryItem{
			{Role: "user", Question: `生成一个hello world网页，点击文字改变文字颜色`},
		},
	}
	_ = os.Chdir(target)
	godotenv.Load()
	var a = fakeChatStream{}
	err := NewChatService(context.Background()).Run(req, &a)
	if err != nil {
		fmt.Println(err)
	}
}
