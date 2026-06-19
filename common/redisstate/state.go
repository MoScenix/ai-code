package redisstate

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type Store struct {
	client redis.Cmdable
	prefix string
}

func NewStore(client redis.Cmdable, prefix string) (*Store, error) {
	if client == nil {
		return nil, fmt.Errorf("redis state store requires client")
	}
	return &Store{
		client: client,
		prefix: strings.Trim(prefix, ":"),
	}, nil
}

func (s *Store) Set(ctx context.Context, key string, value any) error {
	fullKey, err := s.fullKey(key)
	if err != nil {
		return err
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, fullKey, data, 0).Err()
}

func (s *Store) Get(ctx context.Context, key string, out any) (bool, error) {
	fullKey, err := s.fullKey(key)
	if err != nil {
		return false, err
	}

	data, err := s.client.Get(ctx, fullKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}
	if err := json.Unmarshal(data, out); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Store) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	fullKeys := make([]string, 0, len(keys))
	for _, key := range keys {
		fullKey, err := s.fullKey(key)
		if err != nil {
			return err
		}
		fullKeys = append(fullKeys, fullKey)
	}
	return s.client.Del(ctx, fullKeys...).Err()
}

func (s *Store) Expire(ctx context.Context, key string, ttl time.Duration) error {
	fullKey, err := s.fullKey(key)
	if err != nil {
		return err
	}
	return s.client.Expire(ctx, fullKey, ttl).Err()
}

func (s *Store) fullKey(key string) (string, error) {
	key = strings.Trim(key, ":")
	if key == "" {
		return "", fmt.Errorf("redis state key is required")
	}
	if s.prefix == "" {
		return key, nil
	}
	return s.prefix + ":" + key, nil
}
