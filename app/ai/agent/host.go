package agent

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/qwen"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
)

func of[T any](t T) *T {
	return &t
}
func NewChatModel() (*qwen.ChatModel, error) {
	ctx := context.Background()
	apiKey := os.Getenv("DASHSCOPE_API_KEY")
	modelName := os.Getenv("MODEL_NAME")
	cm, err := qwen.NewChatModel(ctx, &qwen.ChatModelConfig{
		BaseURL:     "https://dashscope.aliyuncs.com/compatible-mode/v1",
		APIKey:      apiKey,
		Timeout:     0,
		Model:       modelName,
		MaxTokens:   of(2048),
		Temperature: of(float32(0.7)),
		TopP:        of(float32(0.7)),
	})
	if err != nil {
		log.Fatal(err)
	}
	return cm, err
}
func NewHost(ctx context.Context) *react.Agent {
	cm, err := NewChatModel()
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
	AllTools := compose.ToolsNodeConfig{
		Tools: specialists,
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
