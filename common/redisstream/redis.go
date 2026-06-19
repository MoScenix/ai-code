package redisstream

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client redis.Cmdable
	prefix string
	ttl    time.Duration
}

type Option func(*RedisStore)

func WithTTL(ttl time.Duration) Option {
	return func(s *RedisStore) {
		s.ttl = ttl
	}
}

func NewRedisStore(client redis.Cmdable, prefix string, opts ...Option) (*RedisStore, error) {
	if client == nil {
		return nil, fmt.Errorf("redis stream store requires client")
	}

	store := &RedisStore{
		client: client,
		prefix: strings.Trim(prefix, ":"),
	}
	for _, opt := range opts {
		opt(store)
	}
	return store, nil
}

func (s *RedisStore) Add(ctx context.Context, key string, value any) (string, error) {
	fullKey, err := s.fullKey(key)
	if err != nil {
		return "", err
	}

	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	id, err := s.client.XAdd(ctx, &redis.XAddArgs{
		Stream: fullKey,
		Values: map[string]any{
			DataField: string(data),
		},
	}).Result()
	if err != nil {
		return "", err
	}

	if s.ttl > 0 {
		if err := s.client.Expire(ctx, fullKey, s.ttl).Err(); err != nil {
			return "", err
		}
	}
	return id, nil
}

func (s *RedisStore) Read(ctx context.Context, key string, afterID string, opts ReadOptions) ([]Message, error) {
	fullKey, err := s.fullKey(key)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(afterID) == "" {
		afterID = "$"
	}
	if opts.Count <= 0 {
		opts.Count = 1
	}

	streams, err := s.client.XRead(ctx, &redis.XReadArgs{
		Streams: []string{fullKey, afterID},
		Block:   opts.Block,
		Count:   opts.Count,
	}).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var messages []Message
	for _, stream := range streams {
		for _, msg := range stream.Messages {
			data, err := messageData(msg)
			if err != nil {
				return nil, err
			}
			messages = append(messages, Message{
				ID:   msg.ID,
				Data: data,
			})
		}
	}
	return messages, nil
}

func (s *RedisStore) Del(ctx context.Context, key string) error {
	fullKey, err := s.fullKey(key)
	if err != nil {
		return err
	}
	return s.client.Del(ctx, fullKey).Err()
}

func (s *RedisStore) Expire(ctx context.Context, key string, ttl time.Duration) error {
	fullKey, err := s.fullKey(key)
	if err != nil {
		return err
	}
	return s.client.Expire(ctx, fullKey, ttl).Err()
}

func (s *RedisStore) fullKey(key string) (string, error) {
	key = strings.Trim(key, ":")
	if key == "" {
		return "", fmt.Errorf("redis stream key is required")
	}
	if s.prefix == "" {
		return key, nil
	}
	return s.prefix + ":" + key, nil
}

func messageData(msg redis.XMessage) (json.RawMessage, error) {
	value, ok := msg.Values[DataField]
	if !ok {
		return nil, fmt.Errorf("redis stream message %s missing %q field", msg.ID, DataField)
	}

	switch v := value.(type) {
	case string:
		return json.RawMessage(v), nil
	case []byte:
		return json.RawMessage(v), nil
	default:
		return nil, fmt.Errorf("redis stream message %s %q field has unsupported type %T", msg.ID, DataField, value)
	}
}
