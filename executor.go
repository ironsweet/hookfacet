package hookfacet

import (
	"context"
	"time"
)

type Executor struct {
	H Handler
}

type DispatchResult struct {
	Decision Decision
	Err      error
}

func (e *Executor) Dispatch(ctx context.Context, ev Event, st *State) (Decision, error) {
	done := make(chan *DispatchResult)
	go func() {
		d, err := e.H.OnEvent(ctx, ev, st)
		done <- &DispatchResult{Decision: d, Err: err}
	}()

	timeout := 30 * time.Second
	if e.H.Timeout() > 0 {
		timeout = e.H.Timeout()
	}

	select {
	case res := <-done:
		if e.H.Mode(ev.Kind) == ModeInspect { // inspect mode ignores return value
			return Allow(), res.Err
		}

		if res.Err != nil && e.H.FailureMode() == FailClosed {
			return Block("hook error: " + res.Err.Error()), nil
		}
		return res.Decision, nil
	case <-time.After(timeout):
		if e.H.FailureMode() == FailOpen {
			return Allow(), nil
		}
		return Block("hook timeout"), nil
	}
}
