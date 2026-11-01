package hookfacet

import (
	"context"
	"time"
)

type EventKind string

const (
	KindUserInput EventKind = "user_input"
	KindTurnEnd   EventKind = "turn_end"
)

type Event struct {
	Kind EventKind
}

type HookMode string

const (
	ModeInspect HookMode = "inspect"
	ModeDecide  HookMode = "decide"
)

type FailureMode string

const (
	FailOpen   FailureMode = "fail_open"
	FailClosed FailureMode = "fail_closed"
)

type Decision struct {
	Block  bool
	Reason string
}

func Allow() Decision         { return Decision{} }
func Block(r string) Decision { return Decision{Block: true, Reason: r} }

type Handler interface {
	Mode(EventKind) HookMode
	FailureMode() FailureMode
	Timeout() time.Duration
	OnEvent(ctx context.Context, ev Event, st *State) (Decision, error)
}
