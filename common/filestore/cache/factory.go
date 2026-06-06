package cache

import "github.com/MoScenix/ai-code/common/filestore/base"

func NewStore(actual base.Store) (CacheStore, error) {
	return NewLocalCacheStore(actual)
}
