package agent

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/MoScenix/ai-code/app/ai/llm"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
)

func NewHost(ctx context.Context) *react.Agent {
	fmt.Printf("[ai-agent] host init\n")
	cm, err := llm.NewChatModel(context.Background())
	if err != nil {
		return nil
	}
	var specialists []tool.BaseTool
	PromptsDir, err := ioutil.ReadDir("prompt/specialist")
	for _, p := range PromptsDir {
		if strings.HasSuffix(p.Name(), ".txt") {
			specialistName := strings.TrimSuffix(p.Name(), ".txt")
			specialist, err := NewSpecialist(specialistName)
			if err != nil {
				log.Fatal(err)
			}
			specialists = append(specialists, specialist)
		}
	}
	fmt.Printf("[ai-agent] host ready specialists=%d\n", len(specialists))
	AllTools := compose.ToolsNodeConfig{
		Tools: specialists,
	}
	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: cm,
		ToolsConfig:      AllTools,
		//StreamToolCallChecker: toolCallChecker,
		MaxStep: 12,
	})
	if err != nil {
		log.Fatal(err)
	}
	return agent
}
