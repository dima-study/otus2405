package workerpool

import "context"

type (
	// WorkerJob represents a worker job which could be canceled by ctx.
	//
	// WorkerJob func must have a proper handling for canceled context.
	WorkerJob func(ctx context.Context)

	WorkerPool interface {
		// AddWorker adds new worker to pool. Use ctx to cancel worker and its jobs.
		AddWorker(ctx context.Context)

		// AddJob adds the job to job queue.
		AddJob(job WorkerJob)
	}
)
