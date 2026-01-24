package agent

import (
	"context"
	"os"

	"github.com/MoScenix/ai-code/app/ai/chat"
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
		return SpecialistResult{
			Error: "specialist generate failed: " + err.Error(),
		}, err
	}

	return SpecialistResult{
		Content: Result.Content,
	}, nil
}
func NewSpecialist(Name string) (tool.InvokableTool, error) {
	prompt, err := os.ReadFile("prompt/specialist/" + Name + ".prompt")
	if err != nil {
		return nil, err
	}
	return utils.InferTool(
		Name,
		string(prompt),
		func(ctx context.Context, params *SpecialistParams) (SpecialistResult, error) {
			ctx = context.WithValue(ctx, "specialist", Name)
			return AgentSpecialist(ctx, params)
		})
}
