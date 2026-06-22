package concurrency

import (
	"context"
	"sync"
)

// Supervisor tracks all active goroutines by name to monitor leaks and activity.
type Supervisor struct {
	mu     sync.Mutex
	active map[string]int // name -> count
}

// NewSupervisor creates and initializes a new Supervisor.
func NewSupervisor() *Supervisor {
	return &Supervisor{
		active: make(map[string]int),
	}
}

// Go starts a named goroutine and tracks its lifecycle.
func (s *Supervisor) Go(ctx context.Context, name string, fn func()) {
	s.mu.Lock()
	s.active[name]++
	s.mu.Unlock()

	go func() {
		defer func() {
			s.mu.Lock()
			s.active[name]--
			if s.active[name] == 0 {
				delete(s.active, name) // clean up empty entries
			}
			s.mu.Unlock()
		}()
		// Check context before execution
		select {
		case <-ctx.Done():
			return
		default:
			fn()
		}
	}()
}

// ActiveCount returns a copy of the active goroutine count map.
func (s *Supervisor) ActiveCount() map[string]int {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make(map[string]int, len(s.active))
	for k, v := range s.active {
		out[k] = v
	}
	return out
}
