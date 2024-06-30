package myworkerpool

import (
	"context"
	"sync"

	"github.com/dima-study/otus2405/hw05_parallel_execution/workerpool"
)

// myWorkerPool implements thread-safe workerpool.WorkerPool interface.
//
// Workers should be added first with AddWorker.
// Any job added by AddJob before AddWorker (or after cancel all workers) will be canceled.
type myWorkerPool struct {
	// jobs represents a job queue.
	//
	// It is nil when there are no available workers.
	// The jobs channel has associated ctrl channel.
	jobs chan workerpool.WorkerJob

	// ctrl is a control-channel for job queue.
	//
	// Once ctrl is closed, all accepted and not-handled jobs for associated job queue will be canceled.
	ctrl chan struct{}

	// numWorkers holds number of available workers.
	//
	// Once numWorkers reaches 0 (means "no available workers"),
	// jobs is set to nil and ctrl is scheduled to close.
	numWorkers int

	mu sync.Mutex
}

func New() workerpool.WorkerPool {
	// control-channel is closed initially: there are no available workers yet
	// and all jobs added by AddJob will be canceled.
	ctrl := make(chan struct{})
	close(ctrl)

	return &myWorkerPool{
		ctrl: ctrl,
	}
}

// AddWorker adds a new worker to the worker pool.
// ctx could be used to stop the worker and its jobs.
func (p *myWorkerPool) AddWorker(ctx context.Context) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Initialize communication channels if needed...
	if p.jobs == nil {
		p.ctrl = make(chan struct{})
		p.jobs = make(chan workerpool.WorkerJob)
	}

	// ...and start a worker for the communication channel.
	p.numWorkers++
	go worker(ctx, p.jobs)

	// Each worker has its own stop-the-worker goroutine.
	go func() {
		<-ctx.Done()

		p.mu.Lock()
		defer p.mu.Unlock()

		p.numWorkers--
		if p.numWorkers == 0 {
			// Once the worker is being canceled and if it is the last worker to cancel
			// we need to cancel all accepted but not-handled jobs.
			//
			// Start a new goroutine to handle to cancel all accepted jobs for current job queue in single thread.
			go p.clearJobQueueSlow(p.ctrl, p.jobs)

			p.jobs = nil
		}
	}()
}

// AddJob tries to add a new job to the worker pool it new goroutine, returns immediately.
// If there are no available workers, the job will be canceled.
func (p *myWorkerPool) AddJob(job workerpool.WorkerJob) {
	go func() {
		// Get jobs channel and its associated control-channel.
		p.mu.Lock()
		ctrl := p.ctrl
		jobs := p.jobs
		p.mu.Unlock()

		// If jobs is nil, so there are no available workers.
		if jobs == nil {
			cancelJob(job)
			return
		}

		// The worker pool has an available worker: accept the job and add it to current job queue.
		select {
		case <-ctrl:
			// Looks like all workers has been canceled here, but the job has not been handled: cancel the job.
			cancelJob(job)
		case jobs <- job:
			// The job has been sent to worker successfully.
		}
	}()
}

// clearJobQueueSlow tries to cancel all accepted and not handled yet job for the provided job queue (jobs).
//
// If there are not accepted jobs to handle it closes control-channel for the provided job queue and returns.
func (p *myWorkerPool) clearJobQueueSlow(ctrl chan struct{}, jobs chan workerpool.WorkerJob) {
	for {
		select {
		case job := <-jobs:
			cancelJob(job)
		default:
			// Close control-channel to cancel accepted but not held jobs.
			close(ctrl)
			return
		}
	}
}

// worker is the worker: it accepts job from jobs channel and tries to run it.
// ctx could be used to stop the worker.
func worker(ctx context.Context, jobs chan workerpool.WorkerJob) {
	for {
		// Priority check if worker should be stopped.
		select {
		case <-ctx.Done():
			return
		default:
		}

		select {
		case job := <-jobs:
			job(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// cancelJob force cancels the provided job: job will be run with canceled context.
func cancelJob(job workerpool.WorkerJob) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	job(ctx)
}
