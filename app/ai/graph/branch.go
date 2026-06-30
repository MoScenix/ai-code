package graph

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/MoScenix/ai-code/app/ai/llm"
	"github.com/MoScenix/ai-code/app/ai/utils"
	"github.com/cloudwego/eino/schema"
)

const (
	designerNode = "Designer"
	coderNode    = "Coder"
)

const selectorPromptPath = "prompt/selector/route.prompt"

type routeDecision struct {
	Route string `json:"route"`
}

// newBranch branch initialization method of node 'start' in graph 'aicode'
func newBranch(ctx context.Context, input map[string]any) (endNode string, err error) {
	history, _ := utils.HistoryMessagesFromContext(ctx)
	if !hasUserMessage(history) {
		return coderNode, nil
	}

	cm, err := llm.NewChatModel(ctx)
	if err != nil {
		return coderNode, nil
	}

	prompt, err := os.ReadFile(selectorPromptPath)
	if err != nil {
		return coderNode, nil
	}

	messages := make([]*schema.Message, 0, len(history)+1)
	messages = append(messages, schema.SystemMessage(string(prompt)))
	messages = append(messages, history...)

	msg, err := cm.Generate(ctx, messages)
	if err != nil {
		return coderNode, nil
	}

	return parseRouteJSON(msg.Content), nil
}

func hasUserMessage(messages []*schema.Message) bool {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i] != nil && messages[i].Role == schema.User && strings.TrimSpace(messages[i].Content) != "" {
			return true
		}
	}
	return false
}

func parseRouteJSON(content string) string {
	content = strings.TrimSpace(content)

	var decision routeDecision
	if err := json.Unmarshal([]byte(content), &decision); err != nil {
		return coderNode
	}
	return normalizeRoute(decision.Route)
}

func normalizeRoute(route string) string {
	switch strings.ToLower(strings.TrimSpace(route)) {
	case "designer":
		return designerNode
	case "coder":
		return coderNode
	default:
		return coderNode
	}
}
