package pool

import (
	"context"
	"sync"
)

type PoolInterface interface {
	StartWorkers()
	Submit(job Job) bool
	Stop()
}

// Job represents a unit of work
type Job func()

// Pool represents a goroutine pool
type Pool struct {
	workers int
	ctx     context.Context
	cancel  context.CancelFunc
	jobs    chan Job
	wg      sync.WaitGroup
}

// NewPool creates a new goroutine pool
func NewPool(workers int, workerChannelBuff int) PoolInterface {
	ctx, cancel := context.WithCancel(context.Background())
	return &Pool{
		workers: workers,
		ctx:     ctx,
		cancel:  cancel,
		jobs:    make(chan Job, workerChannelBuff), // Buffer for jobs
		wg:      sync.WaitGroup{},
	}
}

// Submit submits a job to the pool
func (p *Pool) Submit(job Job) bool {
	// Check if context is cancelled first
	select {
	case <-p.ctx.Done():
		return false
	default:
	}

	select {
	case p.jobs <- job:
		return true
	case <-p.ctx.Done():
		return false
	default:
		return false // Pool is full
	}
}

// stops the pool
func (p *Pool) Stop() {
	p.cancel()
	p.wg.Wait()
}

// StartWorkers starts the workers
func (p *Pool) StartWorkers() {
	p.startWorkers(p.ctx, p.workers)
}

// private methods
func (p *Pool) startWorkers(ctx context.Context, workers int) {
	// let start the workers
	for i := 0; i < workers; i++ {
		p.wg.Add(1)
		go p.worker(ctx)
	}
}

// worker is the worker goroutine
func (p *Pool) worker(ctx context.Context) {
	defer p.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return

		case job, ok := <-p.jobs:
			if !ok {
				return
			}

			// execute the job with panic recovery
			func() {
				defer func() {
					if r := recover(); r != nil {
						// Log panic but don't crash the worker
						// In a real implementation, you might want to log this
					}
				}()
				job()
			}()

			// if the job is done, we return
			if ctx.Err() != nil {
				return
			}
		}
	}
}
