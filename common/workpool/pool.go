package workpool

import (
	"context"
	"sync"
	"time"
)

type Pool struct {
	cfg     Config
	onError ErrorHandler

	ctx    context.Context
	cancel context.CancelFunc

	jobs chan Task
	stop chan struct{}

	mu      sync.Mutex
	wg      sync.WaitGroup
	workers map[int]struct{}
	nextID  int
	running int
	closed  bool
}

func New(cfg Config, opts ...Option) (*Pool, error) {
	cfg = cfg.normalize()
	ctx, cancel := context.WithCancel(context.Background())
	p := &Pool{
		cfg:     cfg,
		ctx:     ctx,
		cancel:  cancel,
		jobs:    make(chan Task, cfg.QueueSize),
		stop:    make(chan struct{}),
		workers: make(map[int]struct{}),
	}
	for _, opt := range opts {
		opt(p)
	}

	p.Resize(cfg.MinWorkers)
	return p, nil
}

func (p *Pool) Submit(ctx context.Context, task Task) error {
	if task == nil {
		return ErrNilTask
	}

	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return ErrPoolClosed
	}
	p.mu.Unlock()
	select {
	case p.jobs <- task:
		p.scaleAfterSubmit()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-p.stop:
		return ErrPoolClosed
	case <-p.ctx.Done():
		return ErrPoolClosed
	}
}

func (p *Pool) Resize(target int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}
	if target < p.cfg.MinWorkers {
		target = p.cfg.MinWorkers
	}
	if target > p.cfg.MaxWorkers {
		target = p.cfg.MaxWorkers
	}

	for len(p.workers) < target {
		p.startWorkerLocked()
	}
}

func (p *Pool) Stop(ctx context.Context) error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	close(p.stop)
	p.mu.Unlock()

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		p.cancel()
		return nil
	case <-ctx.Done():
		p.cancel()
		<-done
		return ctx.Err()
	}
}

func (p *Pool) ShutdownNow() {
	p.cancel()
	p.mu.Lock()
	if !p.closed {
		p.closed = true
		close(p.stop)
	}
	p.mu.Unlock()
	p.wg.Wait()
}

func (p *Pool) Stats() Stats {
	p.mu.Lock()
	defer p.mu.Unlock()

	return Stats{
		Workers: len(p.workers),
		Running: p.running,
		Queued:  len(p.jobs),
		Closed:  p.closed,
	}
}

func (p *Pool) startWorkerLocked() {
	id := p.nextID
	p.nextID++
	p.workers[id] = struct{}{}
	p.wg.Add(1)

	go p.workerLoop(id)
}

func (p *Pool) workerLoop(id int) {
	defer p.wg.Done()
	defer p.removeWorker(id)

	timer := time.NewTimer(p.cfg.IdleTimeout)
	defer timer.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-p.stop:
			p.drainJobs()
			return
		case task, ok := <-p.jobs:
			if !ok {
				return
			}
			resetTimer(timer, p.cfg.IdleTimeout)
			p.runTask(task)
		case <-timer.C:
			if p.tryRetireIdle(id) {
				return
			}
			resetTimer(timer, p.cfg.IdleTimeout)
		}
	}
}

func (p *Pool) drainJobs() {
	for {
		select {
		case <-p.ctx.Done():
			return
		case task := <-p.jobs:
			p.runTask(task)
		default:
			return
		}
	}
}

func (p *Pool) runTask(task Task) {
	select {
	case <-p.ctx.Done():
		return
	default:
	}

	p.markRunning(1)
	err := task.Run(p.ctx)
	p.markRunning(-1)
	if err != nil && p.onError != nil {
		p.onError(p.ctx, task, err)
	}
}

func (p *Pool) markRunning(delta int) {
	p.mu.Lock()
	p.running += delta
	p.mu.Unlock()
}

func (p *Pool) removeWorker(id int) {
	p.mu.Lock()
	delete(p.workers, id)
	p.mu.Unlock()
}

func (p *Pool) tryRetireIdle(id int) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return true
	}
	if len(p.workers) <= p.cfg.MinWorkers {
		return false
	}
	if len(p.jobs) > p.cfg.ScaleDownThreshold {
		return false
	}
	return true
}

func (p *Pool) scaleAfterSubmit() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}
	if len(p.jobs) >= p.cfg.ScaleUpThreshold && len(p.workers) < p.cfg.MaxWorkers {
		p.startWorkerLocked()
	}
}

func resetTimer(timer *time.Timer, d time.Duration) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	timer.Reset(d)
}
