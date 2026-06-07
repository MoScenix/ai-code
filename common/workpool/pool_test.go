package workpool

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type countTask struct {
	done *atomic.Int32
}

func (t countTask) Run(ctx context.Context) error {
	t.done.Add(1)
	return nil
}

type blockTask struct {
	block <-chan struct{}
}

func (t blockTask) Run(ctx context.Context) error {
	select {
	case <-t.block:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type failTask struct {
	err error
}

func (t failTask) Run(ctx context.Context) error {
	return t.err
}

func TestPoolSubmitAndStop(t *testing.T) {
	var done atomic.Int32
	p, err := New(Config{
		MinWorkers:       1,
		MaxWorkers:       2,
		QueueSize:        4,
		ScaleUpThreshold: 2,
		IdleTimeout:      20 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 4; i++ {
		if err := p.Submit(context.Background(), countTask{done: &done}); err != nil {
			t.Fatal(err)
		}
	}

	if err := p.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := done.Load(); got != 4 {
		t.Fatalf("expected 4 completed tasks, got %d", got)
	}
	if err := p.Submit(context.Background(), countTask{done: &done}); !errors.Is(err, ErrPoolClosed) {
		t.Fatalf("expected ErrPoolClosed, got %v", err)
	}
}

func TestPoolScaleUpAndIdleRetire(t *testing.T) {
	block := make(chan struct{})
	p, err := New(Config{
		MinWorkers:       0,
		MaxWorkers:       3,
		QueueSize:        8,
		ScaleUpThreshold: 1,
		IdleTimeout:      10 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		if err := p.Submit(context.Background(), blockTask{block: block}); err != nil {
			t.Fatal(err)
		}
	}
	waitUntil(t, 200*time.Millisecond, func() bool {
		return p.Stats().Workers == 3
	})

	close(block)
	waitUntil(t, 200*time.Millisecond, func() bool {
		stats := p.Stats()
		return stats.Workers == 0 && stats.Running == 0
	})

	if err := p.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestPoolErrorHandler(t *testing.T) {
	want := errors.New("failed")
	gotErr := make(chan error, 1)

	p, err := New(Config{
		MinWorkers: 1,
		MaxWorkers: 1,
		QueueSize:  1,
	}, WithErrorHandler(func(ctx context.Context, task Task, err error) {
		gotErr <- err
	}))
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Submit(context.Background(), failTask{err: want}); err != nil {
		t.Fatal(err)
	}

	select {
	case err := <-gotErr:
		if !errors.Is(err, want) {
			t.Fatalf("expected %v, got %v", want, err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for error handler")
	}

	if err := p.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestPoolTaskCarriesParameters(t *testing.T) {
	var mu sync.Mutex
	values := make([]string, 0, 2)

	type appendTask struct {
		value string
	}
	funcRun := func(task appendTask) Task {
		return TaskFunc(func(ctx context.Context) error {
			mu.Lock()
			defer mu.Unlock()
			values = append(values, task.value)
			return nil
		})
	}

	p, err := New(Config{
		MinWorkers: 1,
		MaxWorkers: 1,
		QueueSize:  2,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Submit(context.Background(), funcRun(appendTask{value: "a"})); err != nil {
		t.Fatal(err)
	}
	if err := p.Submit(context.Background(), funcRun(appendTask{value: "b"})); err != nil {
		t.Fatal(err)
	}
	if err := p.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}

	if len(values) != 2 || values[0] != "a" || values[1] != "b" {
		t.Fatalf("unexpected values: %#v", values)
	}
}

func waitUntil(t *testing.T, timeout time.Duration, check func() bool) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if check() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("condition was not met before timeout")
}
