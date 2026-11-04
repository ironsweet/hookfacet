package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/ironsweet/hookfacet"
	"github.com/stretchr/testify/assert"
)

func TestClaudeCode_Read(t *testing.T) {
	for _, tc := range []struct {
		name    string
		input   string
		want    hookfacet.Event
		wantErr bool
	}{
		{
			name:  "UserPromptSubmit",
			input: `{"session_id":"s1","prompt_id":"p1","hook_event_name":"UserPromptSubmit","prompt":"hello"}`,
			want: hookfacet.Event{
				SessionID: "s1",
				TurnID:    "p1",
				Kind:      hookfacet.KindUserInput,
				Payload: hookfacet.Payload{
					Prompt: "hello",
				},
			},
		},
		{
			name:  "PostToolUse",
			input: `{"session_id":"s1","prompt_id":"p1","hook_event_name":"PostToolUse","tool_name":"read_file","tool_use_id":"t1","tool_input":{"file_path":"/x"}}`,
			want: hookfacet.Event{
				SessionID: "s1",
				TurnID:    "p1",
				Kind:      hookfacet.KindToolResult,
				Payload: hookfacet.Payload{
					ToolName:  "read_file",
					ToolUseID: "t1",
					ToolInput: json.RawMessage(`{"file_path":"/x"}`),
				},
			},
		},
		{
			name:  "Stop",
			input: `{"session_id":"s1","prompt_id":"p1","hook_event_name":"Stop","stop_hook_active":true,"last_assistant_message":"done"}`,
			want: hookfacet.Event{
				SessionID: "s1",
				TurnID:    "p1",
				Kind:      hookfacet.KindTurnEnd,
				Payload: hookfacet.Payload{
					StopHookActive:       true,
					LastAssistantMessage: "done",
				},
			},
		},
		{
			name:    "unknown event",
			input:   `{"session_id":"s1","hook_event_name":"Unknown"}`,
			wantErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := &ClaudeCode{}
			got, err := c.Read(bytes.NewBufferString(tc.input))
			assert.True(t, (err != nil) == tc.wantErr, "Read() error = %v, wantErr %v", err, tc.wantErr)
			if !tc.wantErr {
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestClaudeCode_Write(t *testing.T) {
	d := hookfacet.Decision{Block: true, Reason: "intent not satisfied"}
	var buf bytes.Buffer
	err := (&ClaudeCode{}).Write(&buf, d)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"decision":"block","reason":"intent not satisfied"}`, buf.String())
}
