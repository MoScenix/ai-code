package utils

import (
	"context"
	"sync"
	"sync/atomic"
)

const runtimeStateKey contextKey = "runtime_state"

type RuntimeState struct {
	Buffer *StringBuffer

	mu            sync.RWMutex
	controlCursor string
	cancel        context.CancelFunc
	cancelled     atomic.Bool
}

func NewRuntimeState(cancel context.CancelFunc) *RuntimeState {
	return &RuntimeState{
		Buffer:        &StringBuffer{},
		controlCursor: "0",
		cancel:        cancel,
	}
}

func (s *RuntimeState) Cancel() {
	if s == nil {
		return
	}
	s.cancelled.Store(true)
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *RuntimeState) Stop() {
	if s == nil || s.cancel == nil {
		return
	}
	s.cancel()
}

func (s *RuntimeState) IsCancelled() bool {
	return s != nil && s.cancelled.Load()
}

func (s *RuntimeState) ControlCursor() string {
	if s == nil {
		return ""
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.controlCursor
}

func (s *RuntimeState) SetControlCursor(cursor string) {
	if s == nil || cursor == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.controlCursor = cursor
}

func WithRuntimeState(ctx context.Context, state *RuntimeState) context.Context {
	return context.WithValue(ctx, runtimeStateKey, state)
}

func RuntimeStateFromContext(ctx context.Context) (*RuntimeState, bool) {
	state, ok := ctx.Value(runtimeStateKey).(*RuntimeState)
	return state, ok
}

func IsCancelled(ctx context.Context) bool {
	state, ok := RuntimeStateFromContext(ctx)
	return ok && state.IsCancelled()
}

func CancelRuntime(ctx context.Context) {
	if state, ok := RuntimeStateFromContext(ctx); ok {
		state.Cancel()
		return
	}
	if cancel, ok := CancelFuncFromContext(ctx); ok && cancel != nil {
		cancel()
	}
}

func ControlCursor(ctx context.Context) string {
	state, ok := RuntimeStateFromContext(ctx)
	if !ok {
		return ""
	}
	return state.ControlCursor()
}

func SetControlCursor(ctx context.Context, cursor string) {
	if state, ok := RuntimeStateFromContext(ctx); ok {
		state.SetControlCursor(cursor)
	}
}
