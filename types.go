package hookfacet

import (
	"context"
	"time"
)

type EventKind string

const (
	KindUserInput  EventKind = "user_input"
	KindToolResult EventKind = "tool_result"
	KindTurnEnd    EventKind = "turn_end"
)

type Payload struct {
	// UserPromptSubmit
	Prompt string `json:"prompt,omitempty"`

	// PreToolUse
	ToolName  string `json:"tool_name,omitempty"`
	ToolInput *struct {
		FilePath string `json:"file_path,omitempty"`
		Content  string `json:"content,omitempty"`
	} `json:"tool_input,omitempty"`
	ToolUseId string `json:"tool_use_id,omitempty"`
	// PostToolUse
	ToolResponse *struct {
		FilePath string `json:"file_path,omitempty"`
		Content  string `json:"content,omitempty"`
	} `json:"tool_response,omitempty"`

	// Stop
	StopHookActive       bool   `json:"stop_hook_active,omitempty"`
	LastAssistantMessage string `json:"last_assistant_message,omitempty"`
}

type Event struct {
	SessionID string `json:"session_id"`
	TurnID    string `json:"turn_id,omitempty"`

	Kind    EventKind `json:"kind"`
	Payload Payload   `json:"payload"`
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
	Block  bool   `json:"block,omitempty"`
	Reason string `json:"reason,omitempty"`
}

func Allow() Decision         { return Decision{} }
func Block(r string) Decision { return Decision{Block: true, Reason: r} }

type Handler interface {
	Mode(EventKind) HookMode
	FailureMode() FailureMode
	Timeout() time.Duration
	OnEvent(ctx context.Context, ev Event, st *State) (Decision, error)
}
