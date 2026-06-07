package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/MoScenix/ai-code/common/filestore"
	"github.com/MoScenix/ai-code/common/filestore/base"
)

const (
	cacheSuffix = ".cache"
	metaSuffix  = ".meta.json"
)

type LocalCacheStore struct {
	actual     base.Store
	cacheDir   string
	ttl        time.Duration
	needsFlush bool
	mu         sync.Mutex
}

type objectMeta struct {
	Key       string    `json:"key"`
	Dirty     bool      `json:"dirty"`
	Deleted   bool      `json:"deleted"`
	UpdatedAt time.Time `json:"updated_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type cachedObject struct {
	metaPath  string
	cachePath string
	meta      objectMeta
}

func NewLocalCacheStore(actual base.Store) (*LocalCacheStore, error) {
	conf := filestore.GetConf()
	if err := os.MkdirAll(conf.Cache.CacheDir, 0755); err != nil {
		return nil, err
	}
	return &LocalCacheStore{
		actual:     actual,
		cacheDir:   conf.Cache.CacheDir,
		ttl:        time.Duration(conf.Cache.TTLSeconds) * time.Second,
		needsFlush: conf.Cache.NeedFlush,
	}, nil
}

func (s *LocalCacheStore) Read(ctx context.Context, key string) ([]byte, error) {
	if !s.needsFlush {
		return s.actual.Read(ctx, key)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.Lock()
	obj, err := s.loadObject(key)
	if errors.Is(err, filestore.ErrNotFound) {
		s.mu.Unlock()
		return s.actual.Read(ctx, key)
	}
	if err != nil {
		s.mu.Unlock()
		return nil, err
	}
	if expired(obj.meta) {
		_ = s.removeObjectPath(obj)
		s.mu.Unlock()
		return s.actual.Read(ctx, key)
	}
	if obj.meta.Deleted {
		s.mu.Unlock()
		return nil, filestore.ErrNotFound
	}
	data, err := os.ReadFile(obj.cachePath)
	if errors.Is(err, os.ErrNotExist) {
		_ = s.removeObjectPath(obj)
		s.mu.Unlock()
		return s.actual.Read(ctx, key)
	}
	s.mu.Unlock()
	return data, err
}

func (s *LocalCacheStore) Write(ctx context.Context, key string, data []byte) error {
	if !s.needsFlush {
		if err := s.actual.Write(ctx, key, data); err != nil {
			return err
		}
		return s.removeObject(key)
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	obj := s.objectForKey(key)
	if err := os.MkdirAll(filepath.Dir(obj.cachePath), 0755); err != nil {
		return err
	}
	tmp := obj.cachePath + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	if err := os.Rename(tmp, obj.cachePath); err != nil {
		return err
	}
	now := time.Now()
	obj.meta = objectMeta{
		Key:       cleanKey(key),
		Dirty:     true,
		UpdatedAt: now,
		ExpiresAt: now.Add(s.ttl),
	}
	return s.saveMeta(obj)
}

func (s *LocalCacheStore) Delete(ctx context.Context, key string) error {
	if err := s.actual.Delete(ctx, key); err != nil {
		return err
	}
	return s.removeObject(key)
}

func (s *LocalCacheStore) List(ctx context.Context, prefix string) ([]filestore.ObjectInfo, error) {
	if !s.needsFlush {
		return s.actual.List(ctx, prefix)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result := make(map[string]filestore.ObjectInfo)
	actualInfos, err := s.actual.List(ctx, prefix)
	if err != nil && !errors.Is(err, filestore.ErrNotFound) {
		return nil, err
	}
	for _, info := range actualInfos {
		result[info.Key] = info
	}

	s.mu.Lock()
	objects, err := s.scanObjects()
	if err != nil {
		s.mu.Unlock()
		return nil, err
	}
	for _, obj := range objects {
		if expired(obj.meta) {
			_ = s.removeObjectPath(obj)
			continue
		}
		if obj.meta.Deleted || !obj.meta.Dirty {
			continue
		}
		if child, ok := directChild(prefix, obj.meta.Key); ok {
			info := s.objectInfoForChild(prefix, child, obj)
			result[info.Key] = info
		}
	}
	s.mu.Unlock()

	infos := make([]filestore.ObjectInfo, 0, len(result))
	for _, info := range result {
		infos = append(infos, info)
	}
	sortObjectInfos(infos)
	if len(infos) == 0 && errors.Is(err, filestore.ErrNotFound) {
		return nil, filestore.ErrNotFound
	}
	return infos, nil
}

func (s *LocalCacheStore) Stat(ctx context.Context, key string) (filestore.ObjectInfo, error) {
	if !s.needsFlush {
		return s.actual.Stat(ctx, key)
	}
	if err := ctx.Err(); err != nil {
		return filestore.ObjectInfo{}, err
	}

	s.mu.Lock()
	obj, err := s.loadObject(key)
	if err == nil {
		if expired(obj.meta) {
			_ = s.removeObjectPath(obj)
			s.mu.Unlock()
			return s.actual.Stat(ctx, key)
		}
		info, err := cacheFileInfo(obj)
		s.mu.Unlock()
		return info, err
	}
	if !errors.Is(err, filestore.ErrNotFound) {
		s.mu.Unlock()
		return filestore.ObjectInfo{}, err
	}

	objects, err := s.scanObjects()
	if err != nil {
		s.mu.Unlock()
		return filestore.ObjectInfo{}, err
	}
	clean := cleanKey(key)
	for _, obj := range objects {
		if expired(obj.meta) {
			_ = s.removeObjectPath(obj)
			continue
		}
		if strings.HasPrefix(obj.meta.Key, ensureTrailingSlash(clean)) {
			s.mu.Unlock()
			return filestore.ObjectInfo{
				Key:     clean,
				Name:    filepath.Base(clean),
				IsDir:   true,
				ModTime: obj.meta.UpdatedAt,
			}, nil
		}
	}
	s.mu.Unlock()
	return s.actual.Stat(ctx, key)
}

func (s *LocalCacheStore) Flush(ctx context.Context, path string) error {
	if !s.needsFlush {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	path = cleanKey(path)
	s.mu.Lock()
	objects, err := s.scanObjects()
	if err != nil {
		s.mu.Unlock()
		return err
	}
	var targets []cachedObject
	for _, obj := range objects {
		if !inScope(path, obj.meta.Key) {
			continue
		}
		if expired(obj.meta) {
			_ = s.removeObjectPath(obj)
			s.mu.Unlock()
			return filestore.ErrCacheExpired
		}
		if obj.meta.Dirty && !obj.meta.Deleted {
			targets = append(targets, obj)
		}
	}
	sort.Slice(targets, func(i int, j int) bool {
		return targets[i].meta.Key < targets[j].meta.Key
	})
	s.mu.Unlock()

	for _, obj := range targets {
		if err := ctx.Err(); err != nil {
			return err
		}
		data, err := os.ReadFile(obj.cachePath)
		if err != nil {
			return err
		}
		if err := s.actual.Write(ctx, obj.meta.Key, data); err != nil {
			return err
		}
		s.mu.Lock()
		if err := s.removeObjectPath(obj); err != nil {
			s.mu.Unlock()
			return err
		}
		s.mu.Unlock()
	}
	return nil
}

func (s *LocalCacheStore) objectForKey(key string) cachedObject {
	sum := sha256.Sum256([]byte(cleanKey(key)))
	hash := hex.EncodeToString(sum[:])
	dir := filepath.Join(s.cacheDir, hash[:2], hash[2:4])
	return cachedObject{
		metaPath:  filepath.Join(dir, hash+metaSuffix),
		cachePath: filepath.Join(dir, hash+cacheSuffix),
		meta: objectMeta{
			Key: cleanKey(key),
		},
	}
}

func (s *LocalCacheStore) loadObject(key string) (cachedObject, error) {
	obj := s.objectForKey(key)
	data, err := os.ReadFile(obj.metaPath)
	if errors.Is(err, os.ErrNotExist) {
		return cachedObject{}, filestore.ErrNotFound
	}
	if err != nil {
		return cachedObject{}, err
	}
	if err := json.Unmarshal(data, &obj.meta); err != nil {
		return cachedObject{}, err
	}
	if obj.meta.Key != cleanKey(key) {
		return cachedObject{}, filestore.ErrInvalidKey
	}
	return obj, nil
}

func (s *LocalCacheStore) saveMeta(obj cachedObject) error {
	if err := os.MkdirAll(filepath.Dir(obj.metaPath), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(obj.meta, "", "  ")
	if err != nil {
		return err
	}
	tmp := obj.metaPath + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, obj.metaPath)
}

func (s *LocalCacheStore) removeObject(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.removeObjectPath(s.objectForKey(key))
}

func (s *LocalCacheStore) removeObjectPath(obj cachedObject) error {
	if err := os.Remove(obj.cachePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.Remove(obj.metaPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (s *LocalCacheStore) scanObjects() ([]cachedObject, error) {
	var objects []cachedObject
	err := filepath.WalkDir(s.cacheDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), metaSuffix) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var meta objectMeta
		if err := json.Unmarshal(data, &meta); err != nil {
			return err
		}
		cachePath := strings.TrimSuffix(path, metaSuffix) + cacheSuffix
		objects = append(objects, cachedObject{
			metaPath:  path,
			cachePath: cachePath,
			meta:      meta,
		})
		return nil
	})
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	return objects, err
}

func (s *LocalCacheStore) objectInfoForChild(prefix string, child string, obj cachedObject) filestore.ObjectInfo {
	key := joinKey(prefix, child)
	if strings.Contains(strings.TrimPrefix(strings.TrimPrefix(obj.meta.Key, cleanKey(prefix)), "/"), "/") {
		return filestore.ObjectInfo{
			Key:     key,
			Name:    child,
			IsDir:   true,
			ModTime: obj.meta.UpdatedAt,
		}
	}
	info, err := cacheFileInfo(obj)
	if err != nil {
		return filestore.ObjectInfo{
			Key:     key,
			Name:    child,
			ModTime: obj.meta.UpdatedAt,
		}
	}
	info.Key = key
	info.Name = child
	return info
}

func cacheFileInfo(obj cachedObject) (filestore.ObjectInfo, error) {
	info, err := os.Stat(obj.cachePath)
	if errors.Is(err, os.ErrNotExist) {
		return filestore.ObjectInfo{}, filestore.ErrNotFound
	}
	if err != nil {
		return filestore.ObjectInfo{}, err
	}
	return filestore.ObjectInfo{
		Key:     obj.meta.Key,
		Name:    filepath.Base(obj.meta.Key),
		IsDir:   false,
		Size:    info.Size(),
		ModTime: obj.meta.UpdatedAt,
	}, nil
}

func expired(meta objectMeta) bool {
	return !meta.ExpiresAt.IsZero() && time.Now().After(meta.ExpiresAt)
}

func cleanKey(key string) string {
	key = filepath.ToSlash(filepath.Clean(strings.TrimSpace(key)))
	if key == "." {
		return ""
	}
	return strings.TrimPrefix(key, "/")
}

func directChild(prefix string, key string) (string, bool) {
	prefix = cleanKey(prefix)
	key = cleanKey(key)
	if prefix != "" {
		if key == prefix || !strings.HasPrefix(key, ensureTrailingSlash(prefix)) {
			return "", false
		}
		key = strings.TrimPrefix(key, ensureTrailingSlash(prefix))
	}
	child, _, _ := strings.Cut(key, "/")
	return child, child != ""
}

func inScope(prefix string, key string) bool {
	prefix = cleanKey(prefix)
	key = cleanKey(key)
	if prefix == "" {
		return true
	}
	return key == prefix || strings.HasPrefix(key, ensureTrailingSlash(prefix))
}

func joinKey(prefix string, child string) string {
	prefix = cleanKey(prefix)
	if prefix == "" {
		return child
	}
	return filepath.ToSlash(filepath.Join(prefix, child))
}

func ensureTrailingSlash(key string) string {
	key = cleanKey(key)
	if key == "" {
		return ""
	}
	if strings.HasSuffix(key, "/") {
		return key
	}
	return key + "/"
}

func sortObjectInfos(infos []filestore.ObjectInfo) {
	sort.Slice(infos, func(i int, j int) bool {
		if infos[i].IsDir != infos[j].IsDir {
			return infos[i].IsDir
		}
		return infos[i].Name < infos[j].Name
	})
}
