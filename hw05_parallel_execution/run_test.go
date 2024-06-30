package hw05parallelexecution

import (
	"errors"
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestRun(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("if were errors in first M tasks, than finished not more N+M tasks", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i := 0; i < tasksCount; i++ {
			err := fmt.Errorf("error from task %d", i)
			tasks = append(tasks, func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				atomic.AddInt32(&runTasksCount, 1)
				return err
			})
		}

		workersCount := 10
		maxErrorsCount := 23
		err := Run(tasks, workersCount, maxErrorsCount)

		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
		require.LessOrEqual(
			t,
			runTasksCount,
			int32(workersCount+maxErrorsCount),
			"extra tasks were started",
		)
	})

	t.Run("tasks without errors", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var sumTime time.Duration

		for i := 0; i < tasksCount; i++ {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
			sumTime += taskSleep

			tasks = append(tasks, func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})
		}

		workersCount := 5
		maxErrorsCount := 1

		start := time.Now()
		err := Run(tasks, workersCount, maxErrorsCount)
		elapsedTime := time.Since(start)
		require.NoError(t, err)

		require.Equal(t, runTasksCount, int32(tasksCount), "not all tasks were completed")
		require.LessOrEqual(t, int64(elapsedTime), int64(sumTime/2), "tasks were run sequentially?")
	})

	t.Run("ignore errors", func(t *testing.T) {
		tests := []struct {
			name       string
			workersNum int
			tasksNum   int
			errorsNum  int
		}{
			{
				name:       "zero errors",
				workersNum: 100,
				tasksNum:   10,
				errorsNum:  0,
			},
			{
				name:       "with errors",
				workersNum: 100,
				tasksNum:   10,
				errorsNum:  5,
			},
		}

		for _, ts := range tests {
			tasks := make([]Task, ts.tasksNum)
			runTasksCount := int32(0)
			errorsNum := int32(ts.errorsNum)
			for i := range ts.tasksNum {
				taskSleep := time.Millisecond * time.Duration(rand.Intn(100))

				tasks[i] = func() error {
					time.Sleep(taskSleep)
					atomic.AddInt32(&runTasksCount, 1)

					if n := atomic.LoadInt32(&errorsNum); n > 0 {
						atomic.AddInt32(&errorsNum, -1)
						return fmt.Errorf("error %d for task %d", ts.errorsNum-int(n), i)
					}

					return nil
				}
			}

			t.Run(ts.name, func(t *testing.T) {
				err := Run(tasks, ts.workersNum, 0)
				require.NoError(t, err)

				require.Equal(t, runTasksCount, int32(ts.tasksNum), "all tasks must be completed")
			})
		}
	})
}
