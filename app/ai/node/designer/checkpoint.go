package designer

import (
	"context"
	"sync"
)

type memoryCheckpointStore struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func newMemoryCheckpointStore() *memoryCheckpointStore {
	return &memoryCheckpointStore{
		data: make(map[string][]byte),
	}
}

func (s *memoryCheckpointStore) Get(_ context.Context, key string) ([]byte, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, ok := s.data[key]
	if !ok {
		return nil, false, nil
	}
	cp := append([]byte(nil), data...)
	return cp, true, nil
}

func (s *memoryCheckpointStore) Set(_ context.Context, key string, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = append([]byte(nil), value...)
	return nil
}

func (s *memoryCheckpointStore) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, key)
	return nil
}
