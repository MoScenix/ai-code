package agent

import (
	"context"
	"fmt"
	"os"

	"github.com/MoScenix/ai-code/app/ai/chat"
	aitools "github.com/MoScenix/ai-code/app/ai/tools"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

type SpecialistParams struct {
	Content string `json:"content" jsonschema:"description=分析后下达的任务内容"`
}
type SpecialistResult struct {
	Error   string `json:"error" jsonschema:"description=错误信息"`
	Content string `json:"content" jsonschema:"description=专家执行后的结果"`
}

func AgentSpecialist(ctx context.Context, params *SpecialistParams) (SpecialistResult, error) {
	Name := ctx.Value("specialist").(string)
	fmt.Printf("[ai-agent] specialist start name=%s task=%s\n", Name, CompactAgentLog(params.Content))
	agent := chat.NewAiAgent(ctx)
	if _, err := os.Stat("prompt/specialist/" + Name + ".txt"); err != nil {

		return SpecialistResult{
			Error: "specialist not found",
		}, err
	}
	prompt, err := os.ReadFile("prompt/specialist/" + Name + ".txt")
	if err != nil {
		return SpecialistResult{
			Error: "specialist prompt not found",
		}, err
	}
	Result, err := agent.Generate(ctx, []*schema.Message{schema.SystemMessage(string(prompt)), schema.UserMessage(params.Content)})
	if err != nil {
		fmt.Printf("[ai-agent] specialist failed name=%s err=%v\n", Name, err)
		return SpecialistResult{
			Error: "specialist generate failed: " + err.Error(),
		}, err
	}
	LogStreamChunk("specialist:"+Name, Result)
	fmt.Printf("[ai-agent] specialist done name=%s content=%s\n", Name, CompactAgentLog(Result.Content))

	return SpecialistResult{
		Content: Result.Content,
	}, nil
}

func RunSpecialist(ctx context.Context, name string, content string) (SpecialistResult, error) {
	ctx = context.WithValue(ctx, "specialist", name)
	return AgentSpecialist(ctx, &SpecialistParams{Content: content})
}

func NewSpecialist(Name string) (tool.InvokableTool, error) {
	prompt, err := os.ReadFile("prompt/specialist/" + Name + ".prompt")
	if err != nil {
		return nil, err
	}
	t, err := utils.InferTool(
		Name,
		string(prompt),
		func(ctx context.Context, params *SpecialistParams) (SpecialistResult, error) {
			ctx = context.WithValue(ctx, "specialist", Name)
			return AgentSpecialist(ctx, params)
		})
	if err != nil {
		return nil, err
	}
	return aitools.WrapWithLogging(Name, t), nil
}
