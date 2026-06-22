package concurrency

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRunGroup_AllSucceed(t *testing.T) {
	ctx := context.Background()
	err := RunGroup(ctx, 5*time.Second,
		func(ctx context.Context) error { return nil },
		func(ctx context.Context) error { return nil },
		func(ctx context.Context) error { return nil },
	)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestRunGroup_FirstError(t *testing.T) {
	ctx := context.Background()
	sentinel := errors.New("test error")
	err := RunGroup(ctx, 5*time.Second,
		func(ctx context.Context) error { return sentinel },
		func(ctx context.Context) error { time.Sleep(100 * time.Millisecond); return nil },
	)
	if !errors.Is(err, sentinel) {
		t.Errorf("Expected sentinel error, got: %v", err)
	}
}

func TestRunGroup_Timeout(t *testing.T) {
	ctx := context.Background()
	err := RunGroup(ctx, 50*time.Millisecond,
		func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(5 * time.Second):
				return nil
			}
		},
	)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

func TestPipe_PassesChunks(t *testing.T) {
	ctx := context.Background()
	in := make(chan Chunk, 3)
	out := make(chan Chunk, 3)

	in <- Chunk{Data: "hello"}
	in <- Chunk{Data: "world"}
	in <- Chunk{Done: true}
	close(in)

	go Pipe(ctx, in, out)

	got := []string{}
	for chunk := range out {
		if chunk.Data != "" {
			got = append(got, chunk.Data)
		}
		if chunk.Done {
			break
		}
	}

	if len(got) != 2 {
		t.Errorf("Expected 2 chunks, got %d: %v", len(got), got)
	}
}
