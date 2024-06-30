package myworkerpool

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestNew(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		wpool := New()
		require.NotNil(t, wpool, "New workerpool must not be nil")

		_, ok := wpool.(*myWorkerPool)
		require.True(t, ok, "New workerpool must be *myWorkerPool instance")
	})
}

func Test_myWorkerPool_AddWorker(t *testing.T) {
	t.Run("add + cancel workers", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		wpool := New()
		myWPool, ok := wpool.(*myWorkerPool)
		require.True(t, ok, "New workerpool must be *myWorkerPool instance")

		ctxFirst, cancelFirst := context.WithCancel(context.Background())
		const workersNumFirst = 5
		for range workersNumFirst {
			wpool.AddWorker(ctxFirst)
		}

		ctxSecond, cancelSecond := context.WithCancel(context.Background())
		const workersNumSecond = 3
		for range workersNumSecond {
			wpool.AddWorker(ctxSecond)
		}

		require.Equal(
			t,
			workersNumFirst+workersNumSecond,
			myWPool.numWorkers,
			"numWorkers must be equal to workersNumFirst+workersNumSecond after AddWorker",
		)

		timeout := time.NewTimer(30 * time.Second)
		numWorkers := myWPool.numWorkers
		cancelFirst()
		for {
			select {
			case <-timeout.C:
				t.Fatal("timeout: cancelFirst")
			default:
			}

			if !myWPool.mu.TryLock() {
				continue
			}

			n := myWPool.numWorkers
			myWPool.mu.Unlock()
			if numWorkers-n == workersNumFirst {
				break
			}
		}
		if !timeout.Stop() {
			<-timeout.C
		}
		require.Equal(
			t,
			workersNumSecond,
			myWPool.numWorkers,
			"numWorkers must be equal to workersNumSecond after cancel the first ctx: only second workers are live",
		)

		timeout = time.NewTimer(30 * time.Second)
		numWorkers = myWPool.numWorkers
		cancelSecond()
		for {
			select {
			case <-timeout.C:
				t.Fatal("timeout: cancelSecond")
			default:
			}

			if !myWPool.mu.TryLock() {
				continue
			}

			n := myWPool.numWorkers
			myWPool.mu.Unlock()
			if numWorkers-n == workersNumSecond {
				break
			}
		}
		if !timeout.Stop() {
			<-timeout.C
		}
		require.Equal(
			t,
			0,
			myWPool.numWorkers,
			"numWorkers must be equal to zero after cancel the second ctx: there are no live workers after canceling all ctxs",
		)

		require.Nil(t, myWPool.jobs, "jobs channel must be nil")
	})
}

func Test_myWorkerPool_AddJob(t *testing.T) {
	t.Run("add worker + add job + cancel workers + add job", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		wpool := New()

		ctx, cancel := context.WithCancel(context.Background())
		const workersNum = 2
		const numJobs = 9
		const expected = 4
		got := int32(0)

		// Add workersNum workers
		for range workersNum {
			wpool.AddWorker(ctx)
		}

		wg := sync.WaitGroup{}
		wg.Add(numJobs)

		jobResultCh := make(chan struct{})
		canceledCh := make(chan struct{})
		once := sync.Once{}

		// Add numJobs jobs
		for range numJobs {
			wpool.AddJob(func(ctx context.Context) {
				defer wg.Done()

				// Priority cancel
				select {
				case <-ctx.Done():
					once.Do(func() {
						close(canceledCh)
					})
					return
				default:
				}

				// Cancel or do the job
				select {
				case <-ctx.Done():
					once.Do(func() {
						close(canceledCh)
					})

					return
				case jobResultCh <- struct{}{}:
					atomic.AddInt32(&got, 1)
				}
			})
		}

		// Read jobs goroutine
		go func() {
			n := 0
			for range jobResultCh {
				n++
				// Cancel workers once reached expected results
				if n == expected {
					cancel()
					<-canceledCh
				} else if n > expected {
					<-canceledCh
				}
			}
		}()

		wg.Wait()
		close(jobResultCh)

		// Add job when there are no live workers
		wg.Add(1)
		wpool.AddJob(func(ctx context.Context) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				return
			default:
				atomic.AddInt32(&got, 1)
			}
		})
		wg.Wait()

		require.Equalf(t, expected, int(got), "only %d jobs should be done", expected)
	})

	t.Run("add job without workers", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		wpool := New()

		wg := sync.WaitGroup{}
		wg.Add(1)

		c := make(chan struct{})
		wpool.AddJob(func(ctx context.Context) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				return
			case <-c:
				t.Fatal("job added, but must be canceled")
			}
		})

		wg.Wait()
	})
}
