package project

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/MoScenix/ai-code/common/filestore"
	"github.com/MoScenix/ai-code/common/filestore/base"
	"github.com/MoScenix/ai-code/common/filestore/cache"
)

type ProjectStore struct {
	projectID string
	files     cache.CacheStore
}

func NewStore(projectID string, files cache.CacheStore) (*ProjectStore, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, filestore.ErrEmptyProjectID
	}
	if invalidRelativePath(projectID) {
		return nil, filestore.ErrInvalidProjectID
	}
	return &ProjectStore{
		projectID: projectID,
		files:     files,
	}, nil
}

func NewDefaultStore(projectID string) (*ProjectStore, error) {
	actual, err := base.NewStore()
	if err != nil {
		return nil, err
	}
	files, err := cache.NewStore(actual)
	if err != nil {
		return nil, err
	}
	return NewStore(projectID, files)
}

func (s *ProjectStore) ReadFile(ctx context.Context, path string) ([]byte, error) {
	key, err := s.key(path, false)
	if err != nil {
		return nil, err
	}
	return s.files.Read(ctx, key)
}

func (s *ProjectStore) WriteFile(ctx context.Context, path string, data []byte) error {
	key, err := s.key(path, false)
	if err != nil {
		return err
	}
	return s.files.Write(ctx, key, data)
}

func (s *ProjectStore) EditFile(ctx context.Context, path string, oldText string, newText string) error {
	if oldText == "" {
		return filestore.ErrTextNotFound
	}
	data, err := s.ReadFile(ctx, path)
	if err != nil {
		return err
	}
	content := string(data)
	if strings.Count(content, oldText) == 0 {
		return filestore.ErrTextNotFound
	}
	if strings.Count(content, oldText) > 1 {
		return filestore.ErrTextNotUnique
	}
	next := strings.Replace(content, oldText, newText, 1)
	return s.WriteFile(ctx, path, []byte(next))
}

func (s *ProjectStore) InsertAfter(ctx context.Context, path string, anchor string, content string) error {
	return s.EditFile(ctx, path, anchor, anchor+content)
}

func (s *ProjectStore) Delete(ctx context.Context, path string) error {
	key, err := s.key(path, false)
	if err != nil {
		return err
	}
	return s.files.Delete(ctx, key)
}

func (s *ProjectStore) List(ctx context.Context, dir string) ([]filestore.ObjectInfo, error) {
	key, err := s.key(dir, true)
	if err != nil {
		return nil, err
	}
	return s.files.List(ctx, key)
}

func (s *ProjectStore) Stat(ctx context.Context, path string) (filestore.ObjectInfo, error) {
	key, err := s.key(path, true)
	if err != nil {
		return filestore.ObjectInfo{}, err
	}
	return s.files.Stat(ctx, key)
}

func (s *ProjectStore) Commit(ctx context.Context) error {
	return s.files.Flush(ctx)
}

func (s *ProjectStore) key(path string, allowRoot bool) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		if !allowRoot {
			return "", filestore.ErrInvalidKey
		}
		return s.projectID, nil
	}
	if invalidRelativePath(path) {
		return "", filestore.ErrInvalidKey
	}
	return filepath.ToSlash(filepath.Join(s.projectID, path)), nil
}

func invalidRelativePath(path string) bool {
	path = strings.TrimSpace(path)
	if path == "" || filepath.IsAbs(path) || strings.HasPrefix(filepath.ToSlash(path), "/") {
		return true
	}
	parts := strings.Split(filepath.ToSlash(path), "/")
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return true
		}
	}
	return false
}
