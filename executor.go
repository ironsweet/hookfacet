package hookfacet

import (
	"context"
	"time"
)

type Executor struct {
	H Handler
}

type dispatchResult struct {
	Decision Decision
	Err      error
}

func (e *Executor) Dispatch(parent context.Context, ev Event, st *State) (Decision, error) {
	timeout := 30 * time.Second
	if e.H.Timeout() > 0 {
		timeout = e.H.Timeout()
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	done := make(chan *dispatchResult, 1) // buffered to avoid goroutine leak if ctx.Done() fires first
	go func() {
		d, err := e.H.OnEvent(ctx, ev, st)
		done <- &dispatchResult{Decision: d, Err: err}
	}()

	select {
	case res := <-done:
		if e.H.Mode(ev.Kind) == ModeInspect { // inspect mode ignores return value
			return Allow(), res.Err
		}

		if res.Err != nil {
			if e.H.FailureMode() == FailOpen {
				return Allow(), nil
			}
			return Block("hook error: " + res.Err.Error()), nil
		}

		return res.Decision, nil
	case <-ctx.Done():
		if e.H.Mode(ev.Kind) == ModeInspect {
			return Allow(), ctx.Err()
		}

		if e.H.FailureMode() == FailOpen {
			return Allow(), nil
		}
		return Block("hook timeout"), nil
	}
}
