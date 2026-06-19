package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestChat_Run(t *testing.T) {
	shareDir := filepath.Join(t.TempDir(), "project")
	confPath := filepath.Join(t.TempDir(), "filestore.yaml")
	content := []byte("ShareDir:\n  share_dir: " + shareDir + "\ncache:\n  cache_dir: " + filepath.Join(shareDir, "cache") + "\n  ttl_seconds: 7200\n  need_flush: true\n")
	if err := os.WriteFile(confPath, content, 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("FILESTORE_CONF_PATH", confPath)

	_, err := NewChatService(context.Background()).Run("demo")
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}
