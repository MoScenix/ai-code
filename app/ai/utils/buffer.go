package utils

import (
	"context"
	"strings"
	"sync"
)

type stringBufferKey struct{}

type StringBuffer struct {
	mu sync.Mutex
	sb strings.Builder
}

func (b *StringBuffer) WriteString(s string) {
	if b == nil || s == "" {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	_, _ = b.sb.WriteString(s)
}

func (b *StringBuffer) String() string {
	if b == nil {
		return ""
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.sb.String()
}

func (b *StringBuffer) Clear() {
	if b == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.sb.Reset()
}

func (b *StringBuffer) SetString(s string) {
	if b == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.sb.Reset()
	_, _ = b.sb.WriteString(s)
}

func WithStringBuffer(ctx context.Context, buffer *StringBuffer) context.Context {
	return context.WithValue(ctx, stringBufferKey{}, buffer)
}

func StringBufferFromContext(ctx context.Context) (*StringBuffer, bool) {
	buffer, ok := ctx.Value(stringBufferKey{}).(*StringBuffer)
	return buffer, ok
}
