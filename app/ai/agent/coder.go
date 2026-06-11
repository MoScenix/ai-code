package agent

import (
	"context"
	"fmt"
	"os"

	"github.com/MoScenix/ai-code/app/ai/llm"
	aitools "github.com/MoScenix/ai-code/app/ai/tools"
	"github.com/MoScenix/ai-code/common/filestore/project"
	"github.com/cloudwego/eino/adk"
	fsmd "github.com/cloudwego/eino/adk/middlewares/filesystem"
)

const coderInstructionPath = "prompt/coder/instruction.prompt"

func NewCoder(ctx context.Context, store project.Store) (*adk.ChatModelAgent, error) {
	if store == nil {
		return nil, fmt.Errorf("coder agent requires project store")
	}

	cm, err := llm.NewChatModel(ctx)
	if err != nil {
		return nil, err
	}

	filesystemMiddleware, err := fsmd.New(ctx, &fsmd.MiddlewareConfig{
		Backend: aitools.NewProjectFilesystemBackend(store),
	})
	if err != nil {
		return nil, err
	}

	instruction, err := os.ReadFile(coderInstructionPath)
	if err != nil {
		return nil, err
	}

	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "coder",
		Description: "Implements code changes using the project filesystem.",
		Instruction: string(instruction),
		Model:       cm,
		Handlers:    []adk.ChatModelAgentMiddleware{filesystemMiddleware},
	})
}
