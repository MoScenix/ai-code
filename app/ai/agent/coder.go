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
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
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

	filesystemBackend := aitools.NewProjectFilesystemBackend(store)
	writeFileTool, err := aitools.NewWriteFileTool()
	if err != nil {
		return nil, err
	}
	editFileTool, err := aitools.NewEditFileTool()
	if err != nil {
		return nil, err
	}
	filesystemMiddleware, err := fsmd.New(ctx, &fsmd.MiddlewareConfig{
		Backend: filesystemBackend,
		WriteFileToolConfig: &fsmd.ToolConfig{
			Desc:       stringPtr("Write a project-relative file. Keep each write_file call small and valid JSON. Do not put large full-page HTML/CSS/JS payloads in one write_file call; split frontend work into separate files such as index.html, styles.css, and scripts.js, or create a short skeleton first and use edit_file for focused changes."),
			CustomTool: writeFileTool,
		},
		EditFileToolConfig: &fsmd.ToolConfig{
			CustomTool: editFileTool,
		},
	})
	if err != nil {
		return nil, err
	}
	searchTool, err := aitools.NewSearchProjectFileTool()
	if err != nil {
		return nil, err
	}

	instruction, err := os.ReadFile(coderInstructionPath)
	if err != nil {
		return nil, err
	}

	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:             "coder",
		Description:      "Implements code changes using the project filesystem.",
		Instruction:      string(instruction),
		Model:            cm,
		Handlers:         []adk.ChatModelAgentMiddleware{filesystemMiddleware},
		ModelRetryConfig: modelRetryConfig(),
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{searchTool},
			},
		},
	})
}

func stringPtr(v string) *string {
	return &v
}
