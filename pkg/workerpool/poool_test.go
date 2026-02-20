/*
Copyright AppsCode Inc. and Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package workerpool

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestWorkerPool_AllTasksSucceed(t *testing.T) {
	wp := NewWorkerPool(context.Background(), 3, 5*time.Second)

	var executed int64
	for range 5 {
		wp.Run(func() error {
			atomic.AddInt64(&executed, 1)
			time.Sleep(100 * time.Millisecond)
			return nil
		})
	}

	err := wp.Wait()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if executed != 5 {
		t.Fatalf("expected executed=5, got %d", executed)
	}

	if wp.Completed() != 5 {
		t.Fatalf("expected completed=5, got %d", wp.Completed())
	}
}

func TestWorkerPool_ErrorCancelsOthers(t *testing.T) {
	wp := NewWorkerPool(context.Background(), 3, 5*time.Second)

	var executed int64
	for i := range 5 {
		wp.Run(func() error {
			if i == 2 {
				return errors.New("task 2 failed")
			}
			time.Sleep(1 * time.Second)
			atomic.AddInt64(&executed, 1)
			return nil
		})
	}

	err := wp.Wait()
	if err == nil || err.Error() != "task 2 failed" {
		t.Fatalf("expected error 'task 2 failed', got %v", err)
	}
	if wp.Completed() < 1 || wp.Completed() > 5 {
		t.Fatalf("unexpected completed count: %d", wp.Completed())
	}
}

func TestWorkerPool_TimeoutStopsTasks(t *testing.T) {
	wp := NewWorkerPool(context.Background(), 2, 500*time.Millisecond)

	var executed int64
	for range 5 {
		wp.Run(func() error {
			select {
			case <-wp.ctx.Done():
				return wp.ctx.Err()
			case <-time.After(2 * time.Second):
				atomic.AddInt64(&executed, 1)
				return nil
			}
		})
	}

	err := wp.Wait()
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context deadline exceeded, got %v", err)
	}

	// should not execute all tasks
	if executed >= 5 {
		t.Fatalf("expected partial execution, got %d", executed)
	}
}

func TestWorkerPool_PanicRecovery(t *testing.T) {
	var panicCaught atomic.Bool

	handler := func(p any) {
		panicCaught.Store(true)
	}

	wp := NewWorkerPool(context.Background(), 1, 2*time.Second, handler)
	wp.Run(func() error {
		panic("something went wrong")
	})

	err := wp.Wait()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !panicCaught.Load() {
		t.Fatal("expected panic handler to be called")
	}
}

func TestWorkerPool_CancelManually(t *testing.T) {
	wp := NewWorkerPool(context.Background(), 2, 5*time.Second)

	var executed int64
	for range 5 {
		wp.Run(func() error {
			select {
			case <-wp.ctx.Done():
				return wp.ctx.Err()
			case <-time.After(1 * time.Second):
				atomic.AddInt64(&executed, 1)
				return nil
			}
		})
	}

	time.Sleep(300 * time.Millisecond)
	wp.Cancel()
	err := wp.Wait()
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func TestWorkerPool_NoNewTaskAfterTimeout(t *testing.T) {
	wp := NewWorkerPool(context.Background(), 1, 500*time.Millisecond)

	wp.Run(func() error {
		time.Sleep(1 * time.Second)
		return nil
	})

	// wait for timeout
	time.Sleep(600 * time.Millisecond)

	// This should NOT run
	var executed int64
	wp.Run(func() error {
		atomic.AddInt64(&executed, 1)
		return nil
	})

	_ = wp.Wait()
	if executed != 0 {
		t.Fatalf("expected no new task executed after timeout, got %d", executed)
	}
}

func TestWorkerPool_WorkerLimit(t *testing.T) {
	maxWorkers := 2
	wp := NewWorkerPool(context.Background(), maxWorkers, 2*time.Second)

	var concurrent int64
	var maxConcurrent int64

	for range 5 {
		wp.Run(func() error {
			cur := atomic.AddInt64(&concurrent, 1)
			if cur > int64(maxWorkers) {
				t.Fatalf("concurrent workers exceeded limit: %d > %d", cur, maxWorkers)
			}
			if cur > maxConcurrent {
				atomic.StoreInt64(&maxConcurrent, cur)
			}
			time.Sleep(500 * time.Millisecond)
			atomic.AddInt64(&concurrent, -1)
			return nil
		})
	}

	_ = wp.Wait()
	if maxConcurrent != int64(maxWorkers) {
		t.Fatalf("expected max concurrency = %d, got %d", maxWorkers, maxConcurrent)
	}
}
