package concurrency

import (
	"context"
	"time"

	"golang.org/x/sync/errgroup"
)

// RunGroup executes a collection of functions in parallel, bounded by a timeout.
// If any function returns an error, the group context is cancelled and the first error is returned.
func RunGroup(ctx context.Context, timeout time.Duration, fns ...func(ctx context.Context) error) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	for _, fn := range fns {
		fn := fn
		g.Go(func() error {
			return fn(ctx)
		})
	}
	return g.Wait()
}
