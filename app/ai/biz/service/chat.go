package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	lagent "github.com/MoScenix/ai-code/app/ai/agent"
	"github.com/MoScenix/ai-code/app/ai/utils"
	ai "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/ai"
	"github.com/cloudwego/eino/schema"
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
	fmt.Printf("[ai-chat] start project_id=%s history_count=%d fast_path=%t\n", req.ProjectId, len(req.History), isSimpleWebTask(req))
	if isSimpleWebTask(req) {
		return s.runFastPath(req, stream)
	}
	agent := lagent.NewHost(s.ctx)
	var messages []*schema.Message
	prompt, err := os.ReadFile("prompt/host/host.txt")
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
				fmt.Printf("[ai-chat] stream finished project_id=%s\n", req.ProjectId)
				break
			}
			fmt.Printf("[ai-chat] stream error project_id=%s err=%v\n", req.ProjectId, err)
			return err
		}
		lagent.LogStreamChunk("host", msg)
		res := &ai.AiResp{
			Answer: msg.Content,
		}
		fmt.Printf("[ai-chat] send chunk project_id=%s answer=%s\n", req.ProjectId, lagent.CompactAgentLog(msg.Content))
		if err := stream.Send(res); err != nil {
			return err
		}
	}
	return nil
}

func (s *ChatService) runFastPath(req *ai.AiReq, stream ai.AiService_ChatServer) error {
	latest := latestUserQuestion(req)
	if latest == "" {
		return nil
	}

	_ = stream.Send(&ai.AiResp{
		Answer: "检测到这是简单网页任务，已切换快速生成模式。\n",
	})
	fmt.Printf("[ai-chat] fast_path project_id=%s latest=%s\n", req.ProjectId, lagent.CompactAgentLog(latest))

	task := strings.TrimSpace(`请直接完成这个简单网页任务，优先采用最少文件方案（能单文件就单文件），不要先做架构规划，也不要做额外解释。

用户原始需求：
` + latest + `

本步目标：
1. 直接生成可运行网页代码。
2. 优先只创建 index.html；除非确有必要，再拆分 CSS/JS。
3. 用最少的工具调用完成写文件。
4. 完成后只输出简短结果总结。`)

	result, err := lagent.RunSpecialist(s.ctx, "coder", task)
	if err != nil {
		return err
	}

	if result.Error != "" {
		fmt.Printf("[ai-chat] fast_path specialist failed project_id=%s err=%s\n", req.ProjectId, result.Error)
		return errors.New(result.Error)
	}
	fmt.Printf("[ai-chat] fast_path specialist done project_id=%s content=%s\n", req.ProjectId, lagent.CompactAgentLog(result.Content))

	return stream.Send(&ai.AiResp{
		Answer: result.Content,
	})
}

func latestUserQuestion(req *ai.AiReq) string {
	for _, item := range req.History {
		if item.Role == "user" {
			return strings.TrimSpace(item.Question)
		}
	}
	return ""
}

func isSimpleWebTask(req *ai.AiReq) bool {
	if len(req.History) > 4 {
		return false
	}

	latest := strings.ToLower(latestUserQuestion(req))
	if latest == "" {
		return false
	}

	webHints := []string{
		"网页", "页面", "html", "landing page", "web page", "website", "index.html",
	}
	simpleHints := []string{
		"简单", "小", "单页", "展示", "demo", "hello world", "just", "only", "只要", "一个页面",
	}
	complexHints := []string{
		"后台", "管理系统", "多页面", "登录", "注册", "数据库", "接口", "上传", "支付", "聊天", "复杂",
		"admin", "dashboard", "api", "backend", "database", "auth", "multi-page",
	}

	hasWebHint := false
	for _, hint := range webHints {
		if strings.Contains(latest, strings.ToLower(hint)) {
			hasWebHint = true
			break
		}
	}
	if !hasWebHint {
		return false
	}

	for _, hint := range complexHints {
		if strings.Contains(latest, strings.ToLower(hint)) {
			return false
		}
	}

	for _, hint := range simpleHints {
		if strings.Contains(latest, strings.ToLower(hint)) {
			return true
		}
	}

	return len(latest) < 120
}
