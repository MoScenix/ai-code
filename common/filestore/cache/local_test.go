package cache

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/MoScenix/ai-code/common/filestore"
	"github.com/MoScenix/ai-code/common/filestore/base"
)

func TestLocalCacheStoreFlushPrefix(t *testing.T) {
	ctx := context.Background()
	actual, err := base.NewLocalStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	store := &LocalCacheStore{
		actual:     actual,
		cacheDir:   t.TempDir(),
		ttl:        time.Hour,
		needsFlush: true,
	}

	if err := store.Write(ctx, "project_a/index.html", []byte("a")); err != nil {
		t.Fatal(err)
	}
	if err := store.Write(ctx, "project_b/index.html", []byte("b")); err != nil {
		t.Fatal(err)
	}

	if err := store.Flush(ctx, "project_a"); err != nil {
		t.Fatal(err)
	}

	data, err := actual.Read(ctx, "project_a/index.html")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "a" {
		t.Fatalf("expected project_a flushed, got %q", data)
	}

	_, err = actual.Read(ctx, "project_b/index.html")
	if !errors.Is(err, filestore.ErrNotFound) {
		t.Fatalf("expected project_b to stay unflushed, got %v", err)
	}

	data, err = store.Read(ctx, "project_b/index.html")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "b" {
		t.Fatalf("expected project_b cached content, got %q", data)
	}
}

func TestLocalCacheStoreMirrorsKeyPath(t *testing.T) {
	ctx := context.Background()
	actual, err := base.NewLocalStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	cacheDir := t.TempDir()
	store := &LocalCacheStore{
		actual:     actual,
		cacheDir:   cacheDir,
		ttl:        time.Hour,
		needsFlush: true,
	}

	if err := store.Write(ctx, "project_a/src/main.go", []byte("cached")); err != nil {
		t.Fatal(err)
	}

	cachePath := filepath.Join(cacheDir, "project_a", "src", "main.go")
	data, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "cached" {
		t.Fatalf("expected mirrored cache content, got %q", data)
	}
	if _, err := os.Stat(cachePath + metaSuffix); err != nil {
		t.Fatalf("expected mirrored cache metadata: %v", err)
	}
}

func TestLocalCacheStoreFlushSingleFile(t *testing.T) {
	ctx := context.Background()
	actual, err := base.NewLocalStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	store := &LocalCacheStore{
		actual:     actual,
		cacheDir:   t.TempDir(),
		ttl:        time.Hour,
		needsFlush: true,
	}

	if err := store.Write(ctx, "project_a/index.html", []byte("index")); err != nil {
		t.Fatal(err)
	}
	if err := store.Write(ctx, "project_a/src/main.go", []byte("main")); err != nil {
		t.Fatal(err)
	}

	if err := store.Flush(ctx, "project_a/index.html"); err != nil {
		t.Fatal(err)
	}

	data, err := actual.Read(ctx, "project_a/index.html")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "index" {
		t.Fatalf("expected index.html flushed, got %q", data)
	}

	_, err = actual.Read(ctx, "project_a/src/main.go")
	if !errors.Is(err, filestore.ErrNotFound) {
		t.Fatalf("expected src/main.go to stay unflushed, got %v", err)
	}
}

func TestLocalCacheStoreListCachedOverlay(t *testing.T) {
	ctx := context.Background()
	actual, err := base.NewLocalStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	store := &LocalCacheStore{
		actual:     actual,
		cacheDir:   t.TempDir(),
		ttl:        time.Hour,
		needsFlush: true,
	}

	if err := actual.Write(ctx, "project_a/README.md", []byte("base")); err != nil {
		t.Fatal(err)
	}
	if err := store.Write(ctx, "project_a/src/main.go", []byte("cached")); err != nil {
		t.Fatal(err)
	}

	infos, err := store.List(ctx, "project_a")
	if err != nil {
		t.Fatal(err)
	}
	names := make(map[string]bool, len(infos))
	for _, info := range infos {
		names[info.Name] = info.IsDir
	}
	if _, ok := names["README.md"]; !ok {
		t.Fatalf("expected actual file in list: %#v", infos)
	}
	if isDir, ok := names["src"]; !ok || !isDir {
		t.Fatalf("expected cached directory in list: %#v", infos)
	}
}
