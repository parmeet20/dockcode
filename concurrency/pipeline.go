package concurrency

import (
	"context"
)

// Chunk represents a unit of streaming data, carrying text, error, and terminal status.
type Chunk struct {
	Data string
	Err  error // non-nil signals terminal error
	Done bool  // true on last item, then channel closes
}

// Pipeline represents a generic processor that can transform or handle items flowing through channels.
// Here we define helpers for working with Chunk streams.

// Pipe merges two chunk streams or performs an operation on them in a cancellable context.
func Pipe(ctx context.Context, in <-chan Chunk, out chan<- Chunk) {
	defer close(out)
	for {
		select {
		case <-ctx.Done():
			return
		case chunk, ok := <-in:
			if !ok {
				return
			}
			select {
			case out <- chunk:
			case <-ctx.Done():
				return
			}
			if chunk.Err != nil || chunk.Done {
				return
			}
		}
	}
}
