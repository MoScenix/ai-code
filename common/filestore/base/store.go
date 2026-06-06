package base

import (
	"context"

	"github.com/MoScenix/ai-code/common/filestore"
)

// Store is the shared low-level abstraction for file-like object storage.
type Store interface {
	Read(ctx context.Context, key string) ([]byte, error)
	Write(ctx context.Context, key string, data []byte) error
	Delete(ctx context.Context, key string) error
	List(ctx context.Context, prefix string) ([]filestore.ObjectInfo, error)
	Stat(ctx context.Context, key string) (filestore.ObjectInfo, error)
}
