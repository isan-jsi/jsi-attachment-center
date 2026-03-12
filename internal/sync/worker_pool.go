package sync

import (
	"context"
	"log/slog"
	gosync "sync"
)

type Job func(ctx context.Context) error

type WorkerPool struct {
	workers int
	jobs    chan Job
	wg      gosync.WaitGroup
}

func NewWorkerPool(workers int) *WorkerPool {
	return &WorkerPool{
		workers: workers,
		jobs:    make(chan Job, workers*2),
	}
}

func (p *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(ctx, i)
	}
}

func (p *WorkerPool) Submit(job Job) {
	p.jobs <- job
}

func (p *WorkerPool) Wait() {
	close(p.jobs)
	p.wg.Wait()
}

func (p *WorkerPool) worker(ctx context.Context, id int) {
	defer p.wg.Done()
	for job := range p.jobs {
		select {
		case <-ctx.Done():
			slog.Debug("worker stopping", "worker_id", id, "reason", ctx.Err())
			return
		default:
		}

		if err := job(ctx); err != nil {
			slog.Error("worker job failed", "worker_id", id, "error", err)
		}
	}
}
