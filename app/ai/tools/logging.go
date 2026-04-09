package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

const maxLogFieldLen = 600

type loggingTool struct {
	name  string
	inner tool.InvokableTool
}

func WrapWithLogging(name string, inner tool.InvokableTool) tool.InvokableTool {
	return &loggingTool{
		name:  name,
		inner: inner,
	}
}

func (l *loggingTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return l.inner.Info(ctx)
}

func (l *loggingTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	start := time.Now()
	fmt.Printf("[ai-tool] start tool=%s args=%s\n", l.name, compactLog(argumentsInJSON))

	output, err := l.inner.InvokableRun(ctx, argumentsInJSON, opts...)
	if err != nil {
		fmt.Printf("[ai-tool] failed tool=%s cost=%s args=%s err=%v\n", l.name, time.Since(start), compactLog(argumentsInJSON), err)
		return "", err
	}

	fmt.Printf("[ai-tool] done tool=%s cost=%s output=%s\n", l.name, time.Since(start), compactLog(output))
	return output, nil
}

func compactLog(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	if len(s) > maxLogFieldLen {
		return s[:maxLogFieldLen] + "...(truncated)"
	}
	if s == "" {
		return "<empty>"
	}
	return s
}
