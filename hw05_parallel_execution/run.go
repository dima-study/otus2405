package hw05parallelexecution

import (
	"context"
	"errors"
	"sync"

	"github.com/dima-study/otus2405/hw05_parallel_execution/myworkerpool"
	"github.com/dima-study/otus2405/hw05_parallel_execution/workerpool"
)

var (
	ErrErrorsLimitExceeded = errors.New("errors limit exceeded")
	ErrWorkerJobCanceled   = errors.New("job has been canceled")
)

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
// Run ignores tasks errors if m <= 0.
func Run(tasks []Task, n, m int) error {
	workerPool := myworkerpool.New()

	// ctx is context to manage available workers.
	ctx, cancel := context.WithCancel(context.Background())

	// Stop all workers at the end.
	defer cancel()

	// Start n workers.
	for range n {
		workerPool.AddWorker(ctx)
	}

	// wg is used to wait for all tasks to complete.
	wg := sync.WaitGroup{}
	wg.Add(len(tasks))

	// errorsCh is used to count the number of occurred errors.
	errorsCh := make(chan error)
	for _, task := range tasks {
		// Create a new job for each task.
		// Job sends occurred error to the errorsCh channel:
		//   - ErrWorkerJobCanceled once job is canceled by context
		//   - error (or nil) returned from task execution
		var job workerpool.WorkerJob = func(ctx context.Context) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				errorsCh <- ErrWorkerJobCanceled
				return
			default:
			}

			errorsCh <- task()
		}

		workerPool.AddJob(job)
	}

	go func() {
		// Wait until all jobs done and close the errorsCh channel.
		wg.Wait()
		close(errorsCh)
	}()

	// occurredErrors contains the number of occurred errors.
	occurredErrors := 0
	for err := range errorsCh {
		if err != nil {
			// Cancel all jobs and stop all workers when errors limit is reached.
			occurredErrors++
			if occurredErrors == m {
				cancel()
			}
		}
	}

	if m > 0 && occurredErrors >= m {
		return ErrErrorsLimitExceeded
	}

	return nil
}
