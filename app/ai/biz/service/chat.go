package service

import (
	"context"
	"io"
	"os"

	"github.com/MoScenix/ai-code/app/ai/chat"
	"github.com/MoScenix/ai-code/app/ai/utils"
	ai "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/ai"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/kitex/pkg/klog"
)

type ChatService struct {
	ctx context.Context
}

// NewChatService new ChatService
func NewChatService(ctx context.Context) *ChatService {
	return &ChatService{ctx: ctx}
}
func Reverse[T any](s []T) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func (s *ChatService) Run(req *ai.AiReq, stream ai.AiService_ChatServer) (err error) {
	Reverse(req.History)
	s.ctx = context.WithValue(s.ctx, utils.ProjectRootPath, req.ProjectId)
	agent := chat.NewAiAgent(s.ctx)
	var messages []*schema.Message
	prompt, err := os.ReadFile("prompt/HTML-Prompt.txt")
	if err != nil {
		return err
	}
	messages = append(messages, schema.SystemMessage(string(prompt)))
	for _, m := range req.History {
		if m.Role == "user" {
			messages = append(messages, schema.UserMessage(m.Question))
		} else {
			messages = append(messages, schema.AssistantMessage(m.Question, nil))
		}
	}
	msgReader, err := agent.Stream(s.ctx, messages)
	if err != nil {
		return err
	}
	for {
		msg, err := msgReader.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			klog.Error(err)
			return err
		}
		res := &ai.AiResp{
			Answer: msg.Content,
		}
		if err := stream.Send(res); err != nil {
			return err
		}
	}
	return nil
}
