package tools

import (
	"context"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

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
	return l.inner.InvokableRun(ctx, argumentsInJSON, opts...)
}
