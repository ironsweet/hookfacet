package hookfacet

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type statefulStub struct {
	modeFn  func(EventKind) HookMode
	onEvent func(ctx context.Context, ev Event, st *State) (Decision, error)
}

func (r *statefulStub) Mode(kind EventKind) HookMode { return r.modeFn(kind) }
func (r *statefulStub) FailureMode() FailureMode     { return FailClosed }
func (r *statefulStub) Timeout() time.Duration       { return 30 * time.Second }
func (r *statefulStub) OnEvent(ctx context.Context, ev Event, st *State) (Decision, error) {
	return r.onEvent(ctx, ev, st)
}

var _ Handler = (*statefulStub)(nil)

func TestCrossEventStateAggregation(t *testing.T) {
	type round struct {
		userMessage string
		toolCalls   []string
	}

	fakeHandler := &statefulStub{
		modeFn: func(kind EventKind) HookMode {
			switch kind {
			case KindTurnEnd:
				return ModeDecide
			default:
				return ModeInspect
			}
		},
	}
	fakeHandler.onEvent = func(ctx context.Context, ev Event, st *State) (Decision, error) {
		switch ev.Kind {
		case KindUserInput:
			Set(st, round{userMessage: ev.Payload.Prompt})
			return Allow(), nil
		case KindToolResult:
			data, ok := Get[round](st)
			assert.True(t, ok, "expected round data to exist for tool result event")
			assert.Equal(t, "Hello", data.userMessage)
			data.toolCalls = append(data.toolCalls, ev.Payload.ToolName)
			Set(st, data)
			return Allow(), nil
		case KindTurnEnd:
			data, ok := Get[round](st)
			assert.True(t, ok, "expected round data to exist for turn end event")
			assert.Equal(t, "Hello", data.userMessage)
			assert.Equal(t, []string{"toolA", "toolB"}, data.toolCalls)
			return Block("toolC"), nil
		default:
			t.Errorf("unexpected event kind: %v", ev.Kind)
			return Allow(), errors.New("unexpected event kind")
		}
	}
	r := NewRunner(fakeHandler)
	ctx := t.Context()
	d, err := r.Handle(ctx, Event{SessionID: "session1", Kind: KindUserInput, Payload: Payload{Prompt: "Hello"}})
	assert.NoError(t, err)
	assert.Equal(t, Allow(), d)

	d, err = r.Handle(ctx, Event{SessionID: "session1", Kind: KindToolResult, Payload: Payload{ToolName: "toolA"}})
	assert.NoError(t, err)
	assert.Equal(t, Allow(), d)

	d, err = r.Handle(ctx, Event{SessionID: "session1", Kind: KindToolResult, Payload: Payload{ToolName: "toolB"}})
	assert.NoError(t, err)
	assert.Equal(t, Allow(), d)

	d, err = r.Handle(ctx, Event{SessionID: "session1", Kind: KindTurnEnd})
	assert.NoError(t, err)
	assert.Equal(t, Block("toolC"), d)
}

func TestSessionIsolation(t *testing.T) {
	fakeHandler := &statefulStub{
		modeFn: func(kind EventKind) HookMode { return ModeInspect },
		onEvent: func(ctx context.Context, ev Event, st *State) (Decision, error) {
			switch ev.Kind {
			case KindUserInput:
				v, ok := Get[string](st)
				assert.False(t, ok, "expected no existing state for new session")
				assert.Empty(t, v)
				Set[string](st, ev.SessionID)
			}
			return Allow(), nil
		},
	}
	r := NewRunner(fakeHandler)
	ctx := t.Context()
	d, err := r.Handle(ctx, Event{SessionID: "session1", Kind: KindUserInput, Payload: Payload{Prompt: "Hello"}})
	assert.NoError(t, err)
	assert.Equal(t, Allow(), d)

	d, err = r.Handle(ctx, Event{SessionID: "session2", Kind: KindUserInput, Payload: Payload{Prompt: "Hello"}})
	assert.NoError(t, err)
	assert.Equal(t, Allow(), d)
}

func TestConcurrentSession(t *testing.T) {
	fakeHandler := &statefulStub{
		modeFn: func(kind EventKind) HookMode { return ModeInspect },
		onEvent: func(ctx context.Context, ev Event, st *State) (Decision, error) {
			switch ev.Kind {
			case KindUserInput:
				v, ok := Get[string](st)
				assert.False(t, ok, "expected no existing state for new session")
				assert.Empty(t, v)
				Set[string](st, ev.SessionID)
			}
			return Allow(), nil
		},
	}
	r := NewRunner(fakeHandler)
	ctx := t.Context()
	var wg sync.WaitGroup
	for i := range 10 {
		wg.Go(func() {
			d, err := r.Handle(ctx, Event{SessionID: "session" + strconv.Itoa(i), Kind: KindUserInput, Payload: Payload{Prompt: "Hello"}})
			assert.NoError(t, err)
			assert.Equal(t, Allow(), d)
		})
	}
	wg.Wait()
}
