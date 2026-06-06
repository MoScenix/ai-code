package base

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/MoScenix/ai-code/common/filestore"
)

type LocalStore struct {
	root string
}

func NewLocalStore(root string) (*LocalStore, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil, fmt.Errorf("filestore: local root is empty")
	}

	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(abs, 0755); err != nil {
		return nil, err
	}

	return &LocalStore{root: filepath.Clean(abs)}, nil
}

func (s *LocalStore) Read(ctx context.Context, key string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	target, err := s.resolve(key, false)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(target)
	if errors.Is(err, os.ErrNotExist) {
		return nil, filestore.ErrNotFound
	}
	return data, err
}

func (s *LocalStore) Write(ctx context.Context, key string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	target, err := s.resolve(key, false)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	return os.WriteFile(target, data, 0644)
}

func (s *LocalStore) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	target, err := s.resolve(key, false)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(target); err != nil {
		return err
	}
	return nil
}

func (s *LocalStore) List(ctx context.Context, prefix string) ([]filestore.ObjectInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	target, err := s.resolve(prefix, true)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(target)
	if errors.Is(err, os.ErrNotExist) {
		return nil, filestore.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	infos := make([]filestore.ObjectInfo, 0, len(entries))
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		childKey := cleanKey(filepath.Join(prefix, entry.Name()))
		infos = append(infos, objectInfo(childKey, info))
	}
	return infos, nil
}

func (s *LocalStore) Stat(ctx context.Context, key string) (filestore.ObjectInfo, error) {
	if err := ctx.Err(); err != nil {
		return filestore.ObjectInfo{}, err
	}
	target, err := s.resolve(key, true)
	if err != nil {
		return filestore.ObjectInfo{}, err
	}
	info, err := os.Stat(target)
	if errors.Is(err, os.ErrNotExist) {
		return filestore.ObjectInfo{}, filestore.ErrNotFound
	}
	if err != nil {
		return filestore.ObjectInfo{}, err
	}
	return objectInfo(cleanKey(key), info), nil
}

func (s *LocalStore) resolve(key string, allowRoot bool) (string, error) {
	rawKey := strings.TrimSpace(key)
	if rawKey == ".." || strings.HasPrefix(filepath.ToSlash(rawKey), "../") || filepath.IsAbs(rawKey) || strings.HasPrefix(filepath.ToSlash(rawKey), "/") {
		return "", filestore.ErrInvalidKey
	}

	key = cleanKey(key)
	if key == "" && !allowRoot {
		return "", filestore.ErrInvalidKey
	}
	if key == ".." || strings.HasPrefix(key, "../") {
		return "", filestore.ErrInvalidKey
	}

	target := filepath.Join(s.root, key)
	safeTarget, err := resolveForSafety(target)
	if err != nil {
		return "", err
	}
	if !isSubPath(safeTarget, s.root, allowRoot) {
		return "", filestore.ErrInvalidKey
	}
	return target, nil
}

func objectInfo(key string, info os.FileInfo) filestore.ObjectInfo {
	return filestore.ObjectInfo{
		Key:         key,
		Name:        info.Name(),
		IsDir:       info.IsDir(),
		Size:        info.Size(),
		ModTime:     info.ModTime(),
		ETag:        weakETag(info),
		ContentType: contentType(info),
	}
}

func cleanKey(key string) string {
	key = filepath.ToSlash(filepath.Clean(strings.TrimSpace(key)))
	if key == "." {
		return ""
	}
	return strings.TrimPrefix(key, "/")
}

func contentType(info os.FileInfo) string {
	if info.IsDir() {
		return ""
	}
	if ct := mime.TypeByExtension(filepath.Ext(info.Name())); ct != "" {
		return ct
	}
	return "application/octet-stream"
}

func weakETag(info os.FileInfo) string {
	sum := sha1.Sum([]byte(fmt.Sprintf("%s:%d:%d", info.Name(), info.Size(), info.ModTime().UnixNano())))
	return hex.EncodeToString(sum[:])
}

func isSubPath(child, parent string, allowSame bool) bool {
	child = normAbs(child)
	parent = normAbs(parent)
	if child == parent {
		return allowSame
	}
	return strings.HasPrefix(child, ensureTrailingSep(parent))
}

func normAbs(p string) string {
	p = filepath.Clean(p)
	if real, err := filepath.EvalSymlinks(p); err == nil {
		p = filepath.Clean(real)
	}
	if runtime.GOOS == "windows" {
		p = strings.ToLower(p)
	}
	return p
}

func resolveForSafety(p string) (string, error) {
	p = filepath.Clean(p)
	existing := p
	var missing []string
	for {
		if _, err := os.Lstat(existing); err == nil {
			break
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}

		parent := filepath.Dir(existing)
		if parent == existing {
			return p, nil
		}
		missing = append([]string{filepath.Base(existing)}, missing...)
		existing = parent
	}

	realExisting, err := filepath.EvalSymlinks(existing)
	if err != nil {
		return "", err
	}
	parts := append([]string{realExisting}, missing...)
	return filepath.Join(parts...), nil
}

func ensureTrailingSep(p string) string {
	p = filepath.Clean(p)
	if !strings.HasSuffix(p, string(filepath.Separator)) {
		p += string(filepath.Separator)
	}
	return p
}
