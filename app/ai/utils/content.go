package utils

import (
	"context"

	"github.com/MoScenix/ai-code/common/filestore/project"
	"github.com/MoScenix/ai-code/common/redisstate"
	"github.com/MoScenix/ai-code/common/redisstream"
	"github.com/cloudwego/eino/schema"
)

type contextKey string

const (
	historyMessagesKey contextKey = "history_messages"
	projectIDKey       contextKey = "project_id"
	streamStoreKey     contextKey = "stream_store"
	stateStoreKey      contextKey = "state_store"
	cancelFuncKey      contextKey = "cancel_func"
)

var (
	ProjectRootPath  = "project_root"
	ProjectFileStore = "project_file_store"
)

func WithHistoryMessages(ctx context.Context, messages []*schema.Message) context.Context {
	return context.WithValue(ctx, historyMessagesKey, messages)
}

func HistoryMessagesFromContext(ctx context.Context) ([]*schema.Message, bool) {
	messages, ok := ctx.Value(historyMessagesKey).([]*schema.Message)
	return messages, ok
}

func WithProjectID(ctx context.Context, projectID string) context.Context {
	return context.WithValue(ctx, projectIDKey, projectID)
}

func ProjectIDFromContext(ctx context.Context) (string, bool) {
	projectID, ok := ctx.Value(projectIDKey).(string)
	return projectID, ok
}

func WithProjectStore(ctx context.Context, store project.Store) context.Context {
	return context.WithValue(ctx, ProjectFileStore, store)
}

func ProjectStoreFromContext(ctx context.Context) (project.Store, bool) {
	store, ok := ctx.Value(ProjectFileStore).(project.Store)
	return store, ok
}

func WithStreamStore(ctx context.Context, store redisstream.Store) context.Context {
	return context.WithValue(ctx, streamStoreKey, store)
}

func StreamStoreFromContext(ctx context.Context) (redisstream.Store, bool) {
	store, ok := ctx.Value(streamStoreKey).(redisstream.Store)
	return store, ok
}

func WithStateStore(ctx context.Context, store *redisstate.Store) context.Context {
	return context.WithValue(ctx, stateStoreKey, store)
}

func StateStoreFromContext(ctx context.Context) (*redisstate.Store, bool) {
	store, ok := ctx.Value(stateStoreKey).(*redisstate.Store)
	return store, ok
}

func WithCancelFunc(ctx context.Context, cancel context.CancelFunc) context.Context {
	return context.WithValue(ctx, cancelFuncKey, cancel)
}

func CancelFuncFromContext(ctx context.Context) (context.CancelFunc, bool) {
	cancel, ok := ctx.Value(cancelFuncKey).(context.CancelFunc)
	return cancel, ok
}
