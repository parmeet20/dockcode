package concurrency

import (
	"context"
	"sync"
)

// Pool implements a bounded worker pool to restrict concurrency of execution functions.
type Pool struct {
	sem    chan struct{} // semaphore
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
}

// NewPool initializes a new worker pool bounded to maxConcurrent slots.
func NewPool(ctx context.Context, maxConcurrent int) *Pool {
	ctx2, cancel := context.WithCancel(ctx)
	p := &Pool{
		sem:    make(chan struct{}, maxConcurrent),
		ctx:    ctx2,
		cancel: cancel,
		done:   make(chan struct{}),
	}
	return p
}

// Submit schedules a task to run within the worker pool.
// It blocks until a concurrency slot becomes available or context is cancelled.
func (p *Pool) Submit(fn func(ctx context.Context) error) error {
	select {
	case p.sem <- struct{}{}: // acquire slot
	case <-p.ctx.Done():
		return p.ctx.Err()
	}
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		defer func() { <-p.sem }() // release slot

		_ = fn(p.ctx) // Errors are collected via other abstractions (e.g. errgroup) or ignored here
	}()
	return nil
}

// Wait blocks until all active jobs in the pool are finished.
func (p *Pool) Wait() {
	p.wg.Wait()
}

// Stop cancels the context of the pool and waits for all active tasks to complete.
func (p *Pool) Stop() {
	p.cancel()
	p.wg.Wait()
	select {
	case <-p.done:
		// already closed
	default:
		close(p.done)
	}
}
