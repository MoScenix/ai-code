package redisstream

import (
	"context"
	"encoding/json"
	"time"
)

const (
	DataField = "data"
)

type Store interface {
	Add(ctx context.Context, key string, value any) (string, error)
	Read(ctx context.Context, key string, afterID string, opts ReadOptions) ([]Message, error)
	Del(ctx context.Context, key string) error
	Expire(ctx context.Context, key string, ttl time.Duration) error
}

type ReadOptions struct {
	Block time.Duration
	Count int64
}

type Message struct {
	ID   string
	Data json.RawMessage
}

func Decode[T any](msg Message) (T, error) {
	var value T
	err := json.Unmarshal(msg.Data, &value)
	return value, err
}
