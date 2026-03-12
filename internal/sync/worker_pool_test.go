package sync_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	syncsvc "github.com/jsi/ibs-doc-engine/internal/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkerPool_ConcurrencyLimit(t *testing.T) {
	var maxConcurrent int64
	var current int64

	pool := syncsvc.NewWorkerPool(3)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool.Start(ctx)

	for i := 0; i < 10; i++ {
		pool.Submit(func(ctx context.Context) error {
			c := atomic.AddInt64(&current, 1)
			for {
				old := atomic.LoadInt64(&maxConcurrent)
				if c <= old {
					break
				}
				if atomic.CompareAndSwapInt64(&maxConcurrent, old, c) {
					break
				}
			}
			time.Sleep(50 * time.Millisecond)
			atomic.AddInt64(&current, -1)
			return nil
		})
	}

	pool.Wait()

	assert.LessOrEqual(t, maxConcurrent, int64(3), "concurrency should not exceed pool size")
}

func TestWorkerPool_GracefulShutdown(t *testing.T) {
	var completed int64

	pool := syncsvc.NewWorkerPool(2)

	ctx, cancel := context.WithCancel(context.Background())
	pool.Start(ctx)

	for i := 0; i < 5; i++ {
		pool.Submit(func(ctx context.Context) error {
			time.Sleep(20 * time.Millisecond)
			atomic.AddInt64(&completed, 1)
			return nil
		})
	}

	// Cancel context while jobs are running
	time.Sleep(30 * time.Millisecond)
	cancel()
	pool.Wait()

	require.Greater(t, atomic.LoadInt64(&completed), int64(0), "some jobs should have completed")
}
