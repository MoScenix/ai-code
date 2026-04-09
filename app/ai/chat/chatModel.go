package chat

import (
	"context"
	"errors"
	"io"
	"log"

	"github.com/MoScenix/ai-code/app/ai/llm"
	"github.com/MoScenix/ai-code/app/ai/tools"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

func NewAiAgent(ctx context.Context) *react.Agent {
	cm, err := llm.NewChatModel(context.Background())
	if err != nil {
		return nil
	}
	var InvokableTools []tool.InvokableTool
	InvokableTool, err := tools.NewMkPath()
	if err != nil {
		log.Fatal(err)
	}
	InvokableTools = append(InvokableTools, tools.WrapWithLogging("MkPath", InvokableTool))
	InvokableTool, err = tools.NewWriteFileTool()
	if err != nil {
		log.Fatal(err)
	}
	InvokableTools = append(InvokableTools, tools.WrapWithLogging("WriteFile", InvokableTool))
	InvokableTool, err = tools.NewViewFileTool()
	if err != nil {
		log.Fatal(err)
	}
	InvokableTools = append(InvokableTools, tools.WrapWithLogging("ViewFile", InvokableTool))
	InvokableTool, err = tools.NewDeleteFileTool()
	if err != nil {
		log.Fatal(err)
	}
	InvokableTools = append(InvokableTools, tools.WrapWithLogging("DeleteFile", InvokableTool))
	InvokableTool, err = tools.NewReadDirTool()
	if err != nil {
		log.Fatal(err)
	}
	InvokableTools = append(InvokableTools, tools.WrapWithLogging("ReadDir", InvokableTool))
	var baseTools []tool.BaseTool
	for _, t := range InvokableTools {
		baseTools = append(baseTools, t)
	}
	AllTools := compose.ToolsNodeConfig{
		Tools: baseTools,
	}
	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: cm,
		ToolsConfig:      AllTools,
		//StreamToolCallChecker: toolCallChecker,
		MaxStep: 50,
	})
	if err != nil {
		log.Fatal(err)
	}
	return agent
}
func toolCallChecker(ctx context.Context, sr *schema.StreamReader[*schema.Message]) (bool, error) {
	defer sr.Close()
	for {
		msg, err := sr.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				// finish
				break
			}

			return false, err
		}

		if len(msg.ToolCalls) > 0 {
			return true, nil
		}
	}
	return false, nil
}
