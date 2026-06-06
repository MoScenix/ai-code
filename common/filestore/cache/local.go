package cache

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/MoScenix/ai-code/common/filestore"
	"github.com/MoScenix/ai-code/common/filestore/base"
)

const manifestName = ".manifest.json"

type LocalCacheStore struct {
	actual     base.Store
	cache      base.Store
	manifest   string
	ttl        time.Duration
	needsFlush bool
}

type manifest struct {
	CreatedAt time.Time          `json:"created_at"`
	ExpiresAt time.Time          `json:"expires_at"`
	Files     map[string]fileRef `json:"files"`
}

type fileRef struct {
	UpdatedAt time.Time `json:"updated_at"`
}

func NewLocalCacheStore(actual base.Store) (*LocalCacheStore, error) {
	conf := filestore.GetConf()
	cacheStore, err := base.NewLocalStore(conf.Cache.CacheDir)
	if err != nil {
		return nil, err
	}
	return &LocalCacheStore{
		actual:     actual,
		cache:      cacheStore,
		manifest:   filepath.Join(conf.Cache.CacheDir, manifestName),
		ttl:        time.Duration(conf.Cache.TTLSeconds) * time.Second,
		needsFlush: conf.Cache.NeedFlush,
	}, nil
}

func (s *LocalCacheStore) Read(ctx context.Context, key string) ([]byte, error) {
	if !s.needsFlush {
		return s.actual.Read(ctx, key)
	}
	if err := s.dropExpired(ctx); err != nil {
		return nil, err
	}
	data, err := s.cache.Read(ctx, key)
	if err == nil {
		return data, nil
	}
	if !errors.Is(err, filestore.ErrNotFound) {
		return nil, err
	}
	return s.actual.Read(ctx, key)
}

func (s *LocalCacheStore) Write(ctx context.Context, key string, data []byte) error {
	if !s.needsFlush {
		if err := s.actual.Write(ctx, key, data); err != nil {
			return err
		}
		return s.cache.Delete(ctx, key)
	}

	if err := s.dropExpired(ctx); err != nil {
		return err
	}
	if err := s.cache.Write(ctx, key, data); err != nil {
		return err
	}
	return s.markCached(ctx, key)
}

func (s *LocalCacheStore) Delete(ctx context.Context, key string) error {
	if err := s.actual.Delete(ctx, key); err != nil {
		return err
	}
	if err := s.cache.Delete(ctx, key); err != nil {
		return err
	}
	return s.unmarkCached(ctx, key)
}

func (s *LocalCacheStore) List(ctx context.Context, prefix string) ([]filestore.ObjectInfo, error) {
	if !s.needsFlush {
		return s.actual.List(ctx, prefix)
	}
	if err := s.dropExpired(ctx); err != nil {
		return nil, err
	}
	infos, err := s.cache.List(ctx, prefix)
	infos = filterInternalFiles(infos)
	if err == nil && len(infos) > 0 {
		return infos, nil
	}
	if err != nil && !errors.Is(err, filestore.ErrNotFound) {
		return nil, err
	}
	return s.actual.List(ctx, prefix)
}

func (s *LocalCacheStore) Stat(ctx context.Context, key string) (filestore.ObjectInfo, error) {
	if !s.needsFlush {
		return s.actual.Stat(ctx, key)
	}
	if err := s.dropExpired(ctx); err != nil {
		return filestore.ObjectInfo{}, err
	}
	info, err := s.cache.Stat(ctx, key)
	if err == nil {
		return info, nil
	}
	if !errors.Is(err, filestore.ErrNotFound) {
		return filestore.ObjectInfo{}, err
	}
	return s.actual.Stat(ctx, key)
}

func (s *LocalCacheStore) Flush(ctx context.Context) error {
	if !s.needsFlush {
		return nil
	}

	m, err := s.loadManifest()
	if errors.Is(err, filestore.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if expired(m) {
		return filestore.ErrCacheExpired
	}

	keys := make([]string, 0, len(m.Files))
	for key := range m.Files {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		if err := ctx.Err(); err != nil {
			return err
		}
		data, err := s.cache.Read(ctx, key)
		if err != nil {
			return err
		}
		if err := s.actual.Write(ctx, key, data); err != nil {
			return err
		}
	}
	for _, key := range keys {
		if err := s.cache.Delete(ctx, key); err != nil {
			return err
		}
	}
	return s.clearManifest()
}

func (s *LocalCacheStore) markCached(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m, err := s.loadOrNewManifest()
	if err != nil {
		return err
	}
	now := time.Now()
	m.Files[key] = fileRef{UpdatedAt: now}
	m.ExpiresAt = now.Add(s.ttl)
	return s.saveManifest(m)
}

func (s *LocalCacheStore) unmarkCached(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m, err := s.loadManifest()
	if errors.Is(err, filestore.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	delete(m.Files, key)
	return s.saveManifest(m)
}

func (s *LocalCacheStore) dropExpired(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m, err := s.loadManifest()
	if errors.Is(err, filestore.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if !expired(m) {
		return nil
	}
	for key := range m.Files {
		if err := s.cache.Delete(ctx, key); err != nil {
			return err
		}
	}
	return s.clearManifest()
}

func (s *LocalCacheStore) loadOrNewManifest() (*manifest, error) {
	m, err := s.loadManifest()
	if errors.Is(err, filestore.ErrNotFound) {
		now := time.Now()
		return &manifest{
			CreatedAt: now,
			ExpiresAt: now.Add(s.ttl),
			Files:     make(map[string]fileRef),
		}, nil
	}
	return m, err
}

func (s *LocalCacheStore) loadManifest() (*manifest, error) {
	data, err := os.ReadFile(s.manifest)
	if errors.Is(err, os.ErrNotExist) {
		return nil, filestore.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	m := new(manifest)
	if err := json.Unmarshal(data, m); err != nil {
		return nil, err
	}
	if m.Files == nil {
		m.Files = make(map[string]fileRef)
	}
	return m, nil
}

func (s *LocalCacheStore) saveManifest(m *manifest) error {
	if err := os.MkdirAll(filepath.Dir(s.manifest), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.manifest + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, s.manifest)
}

func (s *LocalCacheStore) clearManifest() error {
	if err := os.Remove(s.manifest); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func expired(m *manifest) bool {
	return !m.ExpiresAt.IsZero() && time.Now().After(m.ExpiresAt)
}

func filterInternalFiles(infos []filestore.ObjectInfo) []filestore.ObjectInfo {
	result := infos[:0]
	for _, info := range infos {
		if info.Name == manifestName {
			continue
		}
		result = append(result, info)
	}
	return result
}
