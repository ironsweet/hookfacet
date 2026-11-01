package hookfacet

import "context"

type Executor struct {
	H Handler
}

func (e *Executor) Dispatch(ctx context.Context, ev Event, st *State) (Decision, error) {
	panic("niy")
}
