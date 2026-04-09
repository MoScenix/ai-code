package llm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/qwen"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

const (
	defaultBaseURL        = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	defaultTimeoutSeconds = 120
	defaultMaxRetries     = 2
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

	modelName := strings.TrimSpace(os.Getenv("MODEL_NAME"))
	if modelName == "" {
		return nil, fmt.Errorf("MODEL_NAME is empty")
	}

	timeout := getEnvInt("DASHSCOPE_TIMEOUT_SECONDS", defaultTimeoutSeconds)
	maxRetries := getEnvInt("DASHSCOPE_MAX_RETRIES", defaultMaxRetries)
	baseURL := getEnv("DASHSCOPE_BASE_URL", defaultBaseURL)

	httpClient := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
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
		BaseURL:     baseURL,
		APIKey:      apiKey,
		HTTPClient:  httpClient,
		Model:       modelName,
		MaxTokens:   intPtr(2048),
		Temperature: float32Ptr(0.7),
		TopP:        float32Ptr(0.7),
	})
	if err != nil {
		return nil, err
	}

	return &retryChatModel{
		inner:      cm,
		maxRetries: maxRetries,
	}, nil
}

func getEnv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		log.Printf("invalid %s=%q, fallback to %d", key, raw, fallback)
		return fallback
	}
	return value
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

		log.Printf("retrying DashScope generate request (%d/%d): %v", attempt+1, r.maxRetries, err)
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

		log.Printf("retrying DashScope stream request (%d/%d): %v", attempt+1, r.maxRetries, err)
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
