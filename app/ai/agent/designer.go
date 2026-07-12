package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/MoScenix/ai-code/app/ai/llm"
	aitools "github.com/MoScenix/ai-code/app/ai/tools"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	toolutils "github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

const designerInstructionPath = "prompt/designer/instruction.prompt"

type AskUserInput struct {
	Questions []AskUserQuestion `json:"questions" jsonschema:"description=Questions that must be answered by the user before design can continue."`
	Context   string            `json:"context,omitempty" jsonschema:"description=Short context explaining why these questions are needed."`
}

type AskUserQuestion struct {
	Question string   `json:"question" jsonschema:"description=The question text shown to the user."`
	Options  []string `json:"options,omitempty" jsonschema:"description=Suggested answer options. The user may choose one or provide custom text."`
}

type DesignerInterruptState struct {
	Questions []AskUserQuestion `json:"questions"`
	Context   string            `json:"context,omitempty"`
}

type DesignerAnswer struct {
	TargetID string         `json:"target_id,omitempty"`
	Content  string         `json:"content,omitempty"`
	Payload  map[string]any `json:"payload,omitempty"`
}

func init() {
	schema.RegisterName[AskUserInput]("ai_ask_user_input_v1")
	schema.RegisterName[AskUserQuestion]("ai_ask_user_question_v1")
	schema.RegisterName[DesignerInterruptState]("ai_designer_interrupt_state_v1")
	schema.RegisterName[DesignerAnswer]("ai_designer_answer_v1")
}

func NewDesigner(ctx context.Context) (*adk.ChatModelAgent, error) {
	cm, err := llm.NewChatModel(ctx)
	if err != nil {
		return nil, err
	}

	askTool, err := toolutils.InferTool[AskUserInput, string](
		"ask_user",
		"Ask the user one or more clarification questions and pause execution until answers are provided.",
		runAskUser,
	)
	if err != nil {
		return nil, err
	}
	searchTool, err := aitools.NewSearchProjectFileTool()
	if err != nil {
		return nil, err
	}

	instruction, err := os.ReadFile(designerInstructionPath)
	if err != nil {
		return nil, err
	}

	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:             "designer",
		Description:      "Clarifies requirements and produces a design plan before coding.",
		Instruction:      string(instruction),
		Model:            cm,
		ModelRetryConfig: modelRetryConfig(),
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{askTool, searchTool},
			},
		},
	})
}

func runAskUser(ctx context.Context, input AskUserInput) (string, error) {
	wasInterrupted, _, stored := tool.GetInterruptState[DesignerInterruptState](ctx)
	if !wasInterrupted {
		return "", tool.StatefulInterrupt(ctx, input, DesignerInterruptState{
			Questions: input.Questions,
			Context:   input.Context,
		})
	}

	isTarget, hasData, data := tool.GetResumeContext[DesignerAnswer](ctx)
	if !isTarget {
		return "", tool.StatefulInterrupt(ctx, AskUserInput{
			Questions: stored.Questions,
			Context:   stored.Context,
		}, stored)
	}
	if !hasData {
		return "", fmt.Errorf("ask_user resumed without answer data")
	}

	answer, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(answer), nil
}
