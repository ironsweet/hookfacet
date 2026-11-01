package hookfacet

import "context"

type Runner struct {
}

func NewRunner(h Handler) *Runner {
	panic("niy")
}

func (r *Runner) Handle(ctx context.Context, ev Event) (Decision, error) {
	panic("niy")
}
