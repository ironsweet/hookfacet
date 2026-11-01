package hookfacet

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type stubHandler struct {
	mode        HookMode
	failureMode FailureMode
	timeout     time.Duration
	onEvent     func(ctx context.Context, ev Event, st *State) (Decision, error)
}

func (s *stubHandler) Mode(EventKind) HookMode  { return s.mode }
func (s *stubHandler) FailureMode() FailureMode { return s.failureMode }
func (s *stubHandler) Timeout() time.Duration   { return s.timeout }
func (s *stubHandler) OnEvent(ctx context.Context, ev Event, st *State) (Decision, error) {
	return s.onEvent(ctx, ev, st)
}

var _ Handler = (*stubHandler)(nil)

func TestInspectModeIgnoreReturnValue(t *testing.T) {
	for _, tc := range []struct {
		name    string
		onEvent func(ctx context.Context, ev Event, st *State) (Decision, error)
		assert  func(t *testing.T, d Decision, err error)
	}{
		{
			name: "return block",
			onEvent: func(ctx context.Context, ev Event, st *State) (Decision, error) {
				return Block("should be ignored"), nil
			},
			assert: func(t *testing.T, d Decision, err error) {
				assert.NoError(t, err)
				assert.False(t, d.Block)
			},
		},
		{
			name: "return error",
			onEvent: func(ctx context.Context, ev Event, st *State) (Decision, error) {
				return Allow(), assert.AnError
			},
			assert: func(t *testing.T, d Decision, err error) {
				assert.Error(t, err)
				assert.False(t, d.Block)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fakeHandler := &stubHandler{
				mode:    ModeInspect,
				onEvent: tc.onEvent,
			}
			e := &Executor{H: fakeHandler}
			d, err := e.Dispatch(t.Context(), Event{Kind: KindUserInput}, NewSessionState())
			tc.assert(t, d, err)
		})
	}
}

func TestDecideModePassBlock(t *testing.T) {
	fakeHandler := &stubHandler{
		mode:    ModeDecide,
		onEvent: func(ctx context.Context, ev Event, st *State) (Decision, error) { return Block("blocked"), nil },
	}
	e := &Executor{H: fakeHandler}
	d, err := e.Dispatch(t.Context(), Event{Kind: KindUserInput}, NewSessionState())
	assert.NoError(t, err)
	assert.True(t, d.Block)
	assert.Equal(t, "blocked", d.Reason)
}

func TestDecideModeReturnAllowOnErrorWithFailOpen(t *testing.T) {
	fakeHandler := &stubHandler{
		mode:        ModeDecide,
		failureMode: FailOpen,
		onEvent:     func(ctx context.Context, ev Event, st *State) (Decision, error) { return Allow(), assert.AnError },
	}
	e := &Executor{H: fakeHandler}
	d, err := e.Dispatch(t.Context(), Event{Kind: KindUserInput}, NewSessionState())
	assert.NoError(t, err)
	assert.False(t, d.Block)
}

func TestDecideModeReturnBlockOnErrorWithFailClosed(t *testing.T) {
	fakeHandler := &stubHandler{
		mode:        ModeDecide,
		failureMode: FailClosed,
		onEvent:     func(ctx context.Context, ev Event, st *State) (Decision, error) { return Allow(), assert.AnError },
	}
	e := &Executor{H: fakeHandler}
	d, err := e.Dispatch(t.Context(), Event{Kind: KindUserInput}, NewSessionState())
	assert.NoError(t, err)
	assert.True(t, d.Block)
	assert.NotEmpty(t, d.Reason)
}

func TestTimeoutPerFailureMode(t *testing.T) {
	for _, tc := range []struct {
		name        string
		failureMode FailureMode
		expectBlock bool
	}{
		{name: "fail open", failureMode: FailOpen, expectBlock: false},
		{name: "fail closed", failureMode: FailClosed, expectBlock: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fakeHandler := &stubHandler{
				mode:        ModeDecide,
				failureMode: tc.failureMode,
				timeout:     10 * time.Millisecond,
				onEvent: func(ctx context.Context, ev Event, st *State) (Decision, error) {
					select {
					case <-time.After(500 * time.Millisecond):
						return Allow(), nil
					case <-ctx.Done():
						return Allow(), ctx.Err()
					}
				},
			}
			e := &Executor{H: fakeHandler}
			start := time.Now()
			d, err := e.Dispatch(t.Context(), Event{Kind: KindUserInput}, NewSessionState())
			elapsed := time.Since(start)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectBlock, d.Block)
			assert.Less(t, elapsed, 50*time.Millisecond)
		})
	}
}
