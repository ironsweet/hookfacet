package hookfacet

import (
	"context"
	"sync"
)

type Runner struct {
	e *Executor

	mu     sync.Mutex
	states map[string]*State
}

func NewRunner(h Handler) *Runner {
	return &Runner{e: &Executor{H: h}, states: make(map[string]*State)}
}

func (r *Runner) Handle(ctx context.Context, ev Event) (Decision, error) {
	r.mu.Lock()
	st, ok := r.states[ev.SessionID]
	if !ok {
		st = NewSessionState()
		r.states[ev.SessionID] = st
	}
	r.mu.Unlock()

	return r.e.Dispatch(ctx, ev, st)
}
