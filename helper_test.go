package hookfacet

import (
	"context"
	"time"
)

type stubHandler struct {
	mode        HookMode
	modeFn      func(EventKind) HookMode
	failureMode FailureMode
	timeout     time.Duration
	onEvent     func(ctx context.Context, ev Event, st *State) (Decision, error)
}

func (s *stubHandler) Mode(kind EventKind) HookMode {
	if s.modeFn != nil {
		return s.modeFn(kind)
	}
	return s.mode
}
func (s *stubHandler) FailureMode() FailureMode { return s.failureMode }
func (s *stubHandler) Timeout() time.Duration   { return s.timeout }
func (s *stubHandler) OnEvent(ctx context.Context, ev Event, st *State) (Decision, error) {
	return s.onEvent(ctx, ev, st)
}

var _ Handler = (*stubHandler)(nil)
