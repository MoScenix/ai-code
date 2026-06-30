package llm

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/MoScenix/ai-code/app/ai/conf"
	"github.com/cloudwego/eino-ext/components/model/qwen"
	"github.com/cloudwego/eino/components/model"
)

func intPtr(v int) *int {
	return &v
}

func float32Ptr(v float32) *float32 {
	return &v
}

func NewChatModel(ctx context.Context) (model.ToolCallingChatModel, error) {
	apiKey := strings.TrimSpace(os.Getenv("DASHSCOPE_API_KEY"))
	if apiKey == "" {
		return nil, fmt.Errorf("DASHSCOPE_API_KEY is empty")
	}

	llmConf := conf.GetConf().LLM

	httpClient := &http.Client{
		Timeout: time.Duration(llmConf.TimeoutSeconds) * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   10,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	return qwen.NewChatModel(ctx, &qwen.ChatModelConfig{
		BaseURL:     llmConf.BaseURL,
		APIKey:      apiKey,
		HTTPClient:  httpClient,
		Model:       llmConf.ModelName,
		MaxTokens:   intPtr(llmConf.MaxTokens),
		Temperature: float32Ptr(llmConf.Temperature),
		TopP:        float32Ptr(llmConf.TopP),
	})
}
