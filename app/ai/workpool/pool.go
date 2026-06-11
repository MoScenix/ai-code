package workpool

import (
	"context"
	"sync"
	"time"

	"github.com/MoScenix/ai-code/app/ai/conf"
	commonpool "github.com/MoScenix/ai-code/common/workpool"
	"github.com/cloudwego/kitex/pkg/klog"
)

var (
	once sync.Once
	pool *commonpool.Pool
	err  error
)

func Get() (*commonpool.Pool, error) {
	once.Do(func() {
		cfg := conf.GetConf().WorkPool
		pool, err = commonpool.New(commonpool.Config{
			MinWorkers:         cfg.MinWorkers,
			MaxWorkers:         cfg.MaxWorkers,
			QueueSize:          cfg.QueueSize,
			ScaleUpThreshold:   cfg.ScaleUpThreshold,
			ScaleDownThreshold: cfg.ScaleDownThreshold,
			IdleTimeout:        time.Duration(cfg.IdleTimeoutSeconds) * time.Second,
		}, commonpool.WithErrorHandler(func(ctx context.Context, task commonpool.Task, err error) {
			klog.Errorf("ai task failed: %v", err)
		}))
	})
	return pool, err
}
