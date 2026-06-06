package cache

import (
	"context"

	"github.com/MoScenix/ai-code/common/filestore"
)

// CacheStore represents a local cached view in front of an actual base store.
// Writes go to the cache when flushing is enabled. Delete is passed through to
// the actual store and also removes any local cached copy.
type CacheStore interface {
	Read(ctx context.Context, key string) ([]byte, error)
	Write(ctx context.Context, key string, data []byte) error
	Delete(ctx context.Context, key string) error
	List(ctx context.Context, prefix string) ([]filestore.ObjectInfo, error)
	Stat(ctx context.Context, key string) (filestore.ObjectInfo, error)
	Flush(ctx context.Context) error
}
