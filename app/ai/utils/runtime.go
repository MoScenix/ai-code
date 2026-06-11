package utils

import (
	"context"
	"sync/atomic"
)

const runtimeStateKey contextKey = "runtime_state"

type RuntimeState struct {
	Buffer *StringBuffer

	cancel    context.CancelFunc
	cancelled atomic.Bool
}

func NewRuntimeState(cancel context.CancelFunc) *RuntimeState {
	return &RuntimeState{
		Buffer: &StringBuffer{},
		cancel: cancel,
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
