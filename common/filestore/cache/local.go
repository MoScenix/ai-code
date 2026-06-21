package cache

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/MoScenix/ai-code/common/filestore"
	"github.com/MoScenix/ai-code/common/filestore/base"
)

const metaSuffix = ".meta.json"

// LocalCacheStore implements CacheStore over a local filesystem cache in front
// of a base store. The cache directory mirrors the remote, so no locks needed.
type LocalCacheStore struct {
	actual     base.Store
	cacheDir   string
	ttl        time.Duration
	needsFlush bool
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
	obj, err := s.loadObject(key)
	if errors.Is(err, filestore.ErrNotFound) {
		return s.actual.Read(ctx, key)
	}
	if err != nil {
		return nil, err
	}
	if expired(obj.meta) {
		_ = s.removeObjectPath(obj)
		return s.actual.Read(ctx, key)
	}
	if obj.meta.Deleted {
		return nil, filestore.ErrNotFound
	}
	data, err := os.ReadFile(obj.cachePath)
	if errors.Is(err, os.ErrNotExist) {
		_ = s.removeObjectPath(obj)
		return s.actual.Read(ctx, key)
	}
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
	obj := s.objectForKey(key)
	if err := s.validateObjectPath(obj); err != nil {
		return err
	}
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
		Key: cleanKey(key), Dirty: true, UpdatedAt: now, ExpiresAt: now.Add(s.ttl),
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
	// Non-recursive scan: only immediate children under prefix.
	// Deeper dirty files are collapsed into directory entries by directChild.
	objects, err := s.scanObjectsUnder(prefix, false)
	if err != nil {
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
			result[joinKey(prefix, child)] = s.childInfo(prefix, child, obj)
		}
	}
	infos := make([]filestore.ObjectInfo, 0, len(result))
	for _, info := range result {
		infos = append(infos, info)
	}
	sort.Slice(infos, func(i, j int) bool {
		if infos[i].IsDir != infos[j].IsDir {
			return infos[i].IsDir
		}
		return infos[i].Name < infos[j].Name
	})
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
	obj, err := s.loadObject(key)
	if err == nil {
		if expired(obj.meta) {
			_ = s.removeObjectPath(obj)
			return s.actual.Stat(ctx, key)
		}
		return cacheFileInfo(obj)
	}
	if !errors.Is(err, filestore.ErrNotFound) {
		return filestore.ObjectInfo{}, err
	}
	if info, ok := s.statDir(key); ok {
		return info, nil
	}
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
	// Recursive walk: find all dirty files under path (including subdirs).
	objects, err := s.scanObjectsUnder(path, true)
	if err != nil {
		return err
	}
	var targets []cachedObject
	for _, obj := range objects {
		if !inScope(path, obj.meta.Key) {
			continue
		}
		if expired(obj.meta) {
			_ = s.removeObjectPath(obj)
			return filestore.ErrCacheExpired
		}
		if obj.meta.Dirty && !obj.meta.Deleted {
			targets = append(targets, obj)
		}
	}
	// If path is a single file (not a directory), scanObjectsUnder won't find
	// it — check the exact path too.
	obj, err := s.loadObject(path)
	if err == nil && !expired(obj.meta) && obj.meta.Dirty && !obj.meta.Deleted {
		targets = append(targets, obj)
	}
	sort.Slice(targets, func(i, j int) bool { return targets[i].meta.Key < targets[j].meta.Key })
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
		if err := s.removeObjectPath(obj); err != nil {
			return err
		}
	}
	return nil
}

// --- scan helpers ---

// scanObjectsUnder finds cached entries under prefix. When recursive is true
// (Flush), walks the full subtree; when false (List), only reads the immediate
// directory and returns subdirectories as lightweight markers.
func (s *LocalCacheStore) scanObjectsUnder(prefix string, recursive bool) ([]cachedObject, error) {
	baseDir := filepath.Join(s.cacheDir, filepath.FromSlash(prefix))
	if recursive {
		return s.scanWalkDir(baseDir)
	}
	return s.scanReadDir(baseDir, prefix)
}

func (s *LocalCacheStore) scanWalkDir(baseDir string) ([]cachedObject, error) {
	var objects []cachedObject
	err := filepath.WalkDir(baseDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), metaSuffix) {
			return nil
		}
		obj, err := s.readMeta(path)
		if err != nil {
			return err
		}
		objects = append(objects, obj)
		return nil
	})
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	return objects, err
}

func (s *LocalCacheStore) scanReadDir(baseDir, prefix string) ([]cachedObject, error) {
	entries, err := os.ReadDir(baseDir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var objects []cachedObject
	for _, entry := range entries {
		fullPath := filepath.Join(baseDir, entry.Name())
		if entry.IsDir() {
			fi, _ := entry.Info()
			objects = append(objects, cachedObject{
				cachePath: fullPath,
				meta: objectMeta{
					Key: cleanKey(filepath.ToSlash(filepath.Join(prefix, entry.Name()))),
					Dirty: true, UpdatedAt: fi.ModTime(),
				},
			})
			continue
		}
		if !strings.HasSuffix(entry.Name(), metaSuffix) {
			continue
		}
		obj, err := s.readMeta(fullPath)
		if err != nil {
			return nil, err
		}
		objects = append(objects, obj)
	}
	return objects, nil
}

// readMeta reads a .meta.json file and returns a populated cachedObject.
func (s *LocalCacheStore) readMeta(path string) (cachedObject, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return cachedObject{}, err
	}
	var meta objectMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return cachedObject{}, err
	}
	cachePath := strings.TrimSuffix(path, metaSuffix)
	if meta.Key == "" {
		if rel, err := filepath.Rel(s.cacheDir, cachePath); err == nil {
			meta.Key = cleanKey(filepath.ToSlash(rel))
		}
	}
	return cachedObject{metaPath: path, cachePath: cachePath, meta: meta}, nil
}

// childInfo builds an ObjectInfo for a child entry under prefix.
// The child is a directory when the key is nested deeper (recursive scan) or
// when cachePath points to an actual directory on disk (non-recursive scan).
func (s *LocalCacheStore) childInfo(prefix, child string, obj cachedObject) filestore.ObjectInfo {
	key := joinKey(prefix, child)
	cleanPrefix := cleanKey(prefix)
	// key has more segments past child → it's a directory in the listing
	if strings.Contains(strings.TrimPrefix(obj.meta.Key, ensureTrailingSlash(cleanPrefix)), "/") {
		return filestore.ObjectInfo{Key: key, Name: child, IsDir: true, ModTime: obj.meta.UpdatedAt}
	}
	if fi, err := os.Stat(obj.cachePath); err == nil && fi.IsDir() {
		return filestore.ObjectInfo{Key: key, Name: child, IsDir: true, ModTime: obj.meta.UpdatedAt}
	}
	return fileObjectInfo(key, child, obj)
}

// statDir checks if key is a directory prefix in the local cache by looking at
// the immediate entries under cacheDir/key.
func (s *LocalCacheStore) statDir(key string) (filestore.ObjectInfo, bool) {
	entries, err := os.ReadDir(filepath.Join(s.cacheDir, filepath.FromSlash(key)))
	if err != nil {
		return filestore.ObjectInfo{}, false
	}
	ck := cleanKey(key)
	var latest time.Time
	for _, e := range entries {
		if e.IsDir() {
			if fi, err := e.Info(); err == nil && fi.ModTime().After(latest) {
				latest = fi.ModTime()
			}
			continue
		}
		if strings.HasSuffix(e.Name(), metaSuffix) {
			obj, rerr := s.readMeta(filepath.Join(s.cacheDir, filepath.FromSlash(key), e.Name()))
			if rerr == nil && !obj.meta.Deleted && !expired(obj.meta) && obj.meta.UpdatedAt.After(latest) {
				latest = obj.meta.UpdatedAt
			}
		}
	}
	if latest.IsZero() {
		return filestore.ObjectInfo{}, false
	}
	return filestore.ObjectInfo{Key: ck, Name: filepath.Base(ck), IsDir: true, ModTime: latest}, true
}

// --- key/path utils ---

func (s *LocalCacheStore) objectForKey(key string) cachedObject {
	key = cleanKey(key)
	cachePath := filepath.Join(s.cacheDir, filepath.FromSlash(key))
	return cachedObject{
		metaPath: cachePath + metaSuffix, cachePath: cachePath,
		meta: objectMeta{Key: key},
	}
}

func (s *LocalCacheStore) validateObjectPath(obj cachedObject) error {
	cacheDir, err := filepath.Abs(s.cacheDir)
	if err != nil {
		return err
	}
	cachePath, err := filepath.Abs(obj.cachePath)
	if err != nil {
		return err
	}
	if cachePath == cacheDir || !strings.HasPrefix(cachePath, cacheDir+string(os.PathSeparator)) {
		return filestore.ErrInvalidKey
	}
	return nil
}

func (s *LocalCacheStore) loadObject(key string) (cachedObject, error) {
	obj := s.objectForKey(key)
	if err := s.validateObjectPath(obj); err != nil {
		return cachedObject{}, err
	}
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
	obj := s.objectForKey(key)
	if err := s.validateObjectPath(obj); err != nil {
		return err
	}
	return s.removeObjectPath(obj)
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

// --- file info ---

func fileObjectInfo(key, name string, obj cachedObject) filestore.ObjectInfo {
	fi, err := os.Stat(obj.cachePath)
	if errors.Is(err, os.ErrNotExist) {
		return filestore.ObjectInfo{Key: key, Name: name, ModTime: obj.meta.UpdatedAt}
	}
	if err != nil {
		return filestore.ObjectInfo{Key: key, Name: name, ModTime: obj.meta.UpdatedAt}
	}
	return filestore.ObjectInfo{
		Key: key, Name: name, Size: fi.Size(), ModTime: obj.meta.UpdatedAt,
	}
}

func cacheFileInfo(obj cachedObject) (filestore.ObjectInfo, error) {
	fi, err := os.Stat(obj.cachePath)
	if errors.Is(err, os.ErrNotExist) {
		return filestore.ObjectInfo{}, filestore.ErrNotFound
	}
	if err != nil {
		return filestore.ObjectInfo{}, err
	}
	return filestore.ObjectInfo{
		Key: obj.meta.Key, Name: filepath.Base(obj.meta.Key),
		Size: fi.Size(), ModTime: obj.meta.UpdatedAt,
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

func directChild(prefix, key string) (string, bool) {
	prefix, key = cleanKey(prefix), cleanKey(key)
	if prefix != "" {
		if key == prefix || !strings.HasPrefix(key, ensureTrailingSlash(prefix)) {
			return "", false
		}
		key = strings.TrimPrefix(key, ensureTrailingSlash(prefix))
	}
	child, _, _ := strings.Cut(key, "/")
	return child, child != ""
}

func inScope(prefix, key string) bool {
	prefix, key = cleanKey(prefix), cleanKey(key)
	if prefix == "" {
		return true
	}
	return key == prefix || strings.HasPrefix(key, ensureTrailingSlash(prefix))
}

func joinKey(prefix, child string) string {
	prefix = cleanKey(prefix)
	if prefix == "" {
		return child
	}
	return filepath.ToSlash(filepath.Join(prefix, child))
}

func ensureTrailingSlash(key string) string {
	key = cleanKey(key)
	if key == "" || strings.HasSuffix(key, "/") {
		return key
	}
	return key + "/"
}
