package concurrency

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestPool_Submit(t *testing.T) {
	ctx := context.Background()
	pool := NewPool(ctx, 3)

	var count int64
	for i := 0; i < 9; i++ {
		if err := pool.Submit(func(ctx context.Context) error {
			atomic.AddInt64(&count, 1)
			time.Sleep(5 * time.Millisecond)
			return nil
		}); err != nil {
			t.Fatalf("Submit failed: %v", err)
		}
	}

	pool.Wait()
	if atomic.LoadInt64(&count) != 9 {
		t.Errorf("Expected 9 tasks executed, got %d", count)
	}
}

func TestPool_Stop(t *testing.T) {
	ctx := context.Background()
	pool := NewPool(ctx, 2)

	submitted := 0
	for i := 0; i < 4; i++ {
		err := pool.Submit(func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		})
		if err == nil {
			submitted++
		}
	}

	pool.Stop() // should cancel in-flight and wait
	// Just verify it doesn't hang
}

func TestPool_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	pool := NewPool(ctx, 2)
	// Saturate pool
	_ = pool.Submit(func(ctx context.Context) error {
		time.Sleep(200 * time.Millisecond)
		return nil
	})
	_ = pool.Submit(func(ctx context.Context) error {
		time.Sleep(200 * time.Millisecond)
		return nil
	})

	cancel()

	err := pool.Submit(func(ctx context.Context) error { return nil })
	if err == nil {
		t.Error("Expected error when submitting to cancelled pool")
	}
	pool.Wait()
}
