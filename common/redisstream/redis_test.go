package redisstream

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestDecode(t *testing.T) {
	type event struct {
		Type    string `json:"type"`
		Content string `json:"content"`
	}

	got, err := Decode[event](Message{
		ID:   "1-0",
		Data: []byte(`{"type":"message","content":"hello"}`),
	})
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if got.Type != "message" || got.Content != "hello" {
		t.Fatalf("Decode() = %+v", got)
	}
}

func TestFullKey(t *testing.T) {
	store, err := NewRedisStore(redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"}), "ai")
	if err != nil {
		t.Fatalf("NewRedisStore() error = %v", err)
	}

	got, err := store.fullKey(":task:1:event:")
	if err != nil {
		t.Fatalf("fullKey() error = %v", err)
	}
	if got != "ai:task:1:event" {
		t.Fatalf("fullKey() = %q", got)
	}
}

func TestReadRejectsEmptyKey(t *testing.T) {
	store, err := NewRedisStore(redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"}), "ai")
	if err != nil {
		t.Fatalf("NewRedisStore() error = %v", err)
	}

	_, err = store.Read(context.Background(), "", "0", ReadOptions{})
	if err == nil {
		t.Fatal("Read() expected error")
	}
}
