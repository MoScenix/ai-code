package project

import (
	"context"

	"github.com/MoScenix/ai-code/common/filestore"
)

type Store interface {
	ReadFile(ctx context.Context, path string) ([]byte, error)
	WriteFile(ctx context.Context, path string, data []byte) error
	EditFile(ctx context.Context, path string, oldText string, newText string) error
	InsertAfter(ctx context.Context, path string, anchor string, content string) error
	Delete(ctx context.Context, path string) error
	List(ctx context.Context, dir string) ([]filestore.ObjectInfo, error)
	Stat(ctx context.Context, path string) (filestore.ObjectInfo, error)
	Commit(ctx context.Context) error
}
