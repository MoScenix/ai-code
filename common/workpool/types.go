package workpool

import (
	"context"
	"errors"
	"time"
)

var (
	ErrPoolClosed = errors.New("workpool is closed")
	ErrNilTask    = errors.New("workpool task is nil")
)

type Task interface {
	Run(ctx context.Context) error
}

type TaskFunc func(ctx context.Context) error

func (f TaskFunc) Run(ctx context.Context) error {
	return f(ctx)
}

type ErrorHandler func(ctx context.Context, task Task, err error)

type Option func(*Pool)

func WithErrorHandler(handler ErrorHandler) Option {
	return func(p *Pool) {
		p.onError = handler
	}
}

type Config struct {
	MinWorkers int
	MaxWorkers int
	QueueSize  int

	ScaleUpThreshold   int
	ScaleDownThreshold int
	IdleTimeout        time.Duration
}

func (c Config) normalize() Config {
	if c.MinWorkers < 0 {
		c.MinWorkers = 0
	}
	if c.MaxWorkers <= 0 {
		c.MaxWorkers = 1
	}
	if c.MinWorkers > c.MaxWorkers {
		c.MinWorkers = c.MaxWorkers
	}
	if c.QueueSize <= 0 {
		c.QueueSize = c.MaxWorkers
	}
	if c.ScaleUpThreshold <= 0 || c.ScaleUpThreshold > c.QueueSize {
		c.ScaleUpThreshold = c.QueueSize
	}
	if c.ScaleDownThreshold < 0 {
		c.ScaleDownThreshold = 0
	}
	if c.ScaleDownThreshold >= c.ScaleUpThreshold {
		c.ScaleDownThreshold = c.ScaleUpThreshold - 1
	}
	if c.IdleTimeout <= 0 {
		c.IdleTimeout = 30 * time.Second
	}
	return c
}

type Stats struct {
	Workers int
	Running int
	Queued  int
	Closed  bool
}
