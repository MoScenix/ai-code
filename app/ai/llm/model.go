package llm

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/MoScenix/ai-code/app/ai/conf"
	"github.com/cloudwego/eino-ext/components/model/qwen"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/kitex/pkg/klog"
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

	cm, err := qwen.NewChatModel(ctx, &qwen.ChatModelConfig{
		BaseURL:     llmConf.BaseURL,
		APIKey:      apiKey,
		HTTPClient:  httpClient,
		Model:       llmConf.ModelName,
		MaxTokens:   intPtr(llmConf.MaxTokens),
		Temperature: float32Ptr(llmConf.Temperature),
		TopP:        float32Ptr(llmConf.TopP),
	})
	if err != nil {
		return nil, err
	}

	return &retryChatModel{
		inner:      cm,
		maxRetries: llmConf.MaxRetries,
	}, nil
}

type retryChatModel struct {
	inner      model.ToolCallingChatModel
	maxRetries int
}

func (r *retryChatModel) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	var lastErr error
	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		out, err := r.inner.Generate(ctx, input, opts...)
		if err == nil {
			return out, nil
		}

		lastErr = err
		if attempt == r.maxRetries || !isRetryableError(err) {
			return nil, err
		}

		klog.CtxWarnf(ctx, "retry dashscope generate: attempt=%d max_retries=%d err=%v", attempt+1, r.maxRetries, err)
		time.Sleep(backoff(attempt + 1))
	}

	return nil, lastErr
}

func (r *retryChatModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	var lastErr error
	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		out, err := r.inner.Stream(ctx, input, opts...)
		if err == nil {
			return out, nil
		}

		lastErr = err
		if attempt == r.maxRetries || !isRetryableError(err) {
			return nil, err
		}

		klog.CtxWarnf(ctx, "retry dashscope stream: attempt=%d max_retries=%d err=%v", attempt+1, r.maxRetries, err)
		time.Sleep(backoff(attempt + 1))
	}

	return nil, lastErr
}

func (r *retryChatModel) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	next, err := r.inner.WithTools(tools)
	if err != nil {
		return nil, err
	}
	return &retryChatModel{
		inner:      next,
		maxRetries: r.maxRetries,
	}, nil
}

func (r *retryChatModel) GetType() string {
	if typed, ok := r.inner.(interface{ GetType() string }); ok {
		return typed.GetType()
	}
	return "RetryChatModel"
}

func backoff(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	delay := time.Duration(attempt) * time.Second
	if delay > 5*time.Second {
		return 5 * time.Second
	}
	return delay
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) && (netErr.Timeout() || netErr.Temporary()) {
		return true
	}

	msg := strings.ToLower(err.Error())
	retryableSubstrings := []string{
		"timeout",
		"timed out",
		"connection reset by peer",
		"broken pipe",
		"unexpected eof",
		"eof",
		"connection refused",
		"server closed idle connection",
		"tls handshake timeout",
		"http2: client connection lost",
		"502 bad gateway",
		"503 service unavailable",
		"504 gateway timeout",
		"too many requests",
		"429",
	}
	for _, part := range retryableSubstrings {
		if strings.Contains(msg, part) {
			return true
		}
	}

	return false
}
