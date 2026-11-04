package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/ironsweet/hookfacet"
)

type ClaudeCodeHookInput struct {
	// Core
	SessionID      string `json:"session_id"`
	PromptID       string `json:"prompt_id,omitempty"`
	TranscriptPath string `json:"transcript_path"`
	CWD            string `json:"cwd"`
	ScratchpadDir  string `json:"scratchpad_dir,omitempty"`
	PermissionMode string `json:"permission_mode"`
	Effort         string `json:"effort"`
	HookEventName  string `json:"hook_event_name"`

	// SessionStart
	Source       string `json:"source,omitempty"`
	Model        string `json:"model,omitempty"`
	AgentType    string `json:"agent_type,omitempty"`
	SessionTitle string `json:"session_title,omitempty"`
	// when source="resume|fork"
	SecondsSinceLastResponse int     `json:"seconds_since_last_response,omitempty"`
	ContextTokens            int     `json:"context_tokens,omitempty"`
	PromptCacheLikelyExpired bool    `json:"prompt_cache_likely_expired,omitempty"`
	EstimatedCacheWriteUsd   float64 `json:"estimated_cache_write_usd,omitempty"`

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

type ClaudeCodeDecision struct {
	Decision string `json:"decision,omitempty"` // "block" or empty
	Reason   string `json:"reason,omitempty"`
}

type ClaudeCode struct {
}

func (c *ClaudeCode) Read(r io.Reader) (hookfacet.Event, error) {
	var hookInput ClaudeCodeHookInput
	err := json.NewDecoder(r).Decode(&hookInput)
	if err != nil {
		return hookfacet.Event{}, err
	}
	switch hookInput.HookEventName {
	case "UserPromptSubmit":
		return hookfacet.Event{
			SessionID: hookInput.SessionID,
			TurnID:    hookInput.PromptID,
			Kind:      hookfacet.KindUserInput,
			Payload: hookfacet.Payload{
				Prompt: hookInput.Prompt,
			},
		}, nil
	case "PostToolUse":
		return hookfacet.Event{
			SessionID: hookInput.SessionID,
			TurnID:    hookInput.ToolUseId,
			Kind:      hookfacet.KindToolResult,
			Payload: hookfacet.Payload{
				ToolName:     hookInput.ToolName,
				ToolInput:    hookInput.ToolInput,
				ToolUseId:    hookInput.ToolUseId,
				ToolResponse: hookInput.ToolResponse,
			},
		}, nil
	case "Stop":
		return hookfacet.Event{
			SessionID: hookInput.SessionID,
			TurnID:    hookInput.PromptID,
			Kind:      hookfacet.KindTurnEnd,
			Payload: hookfacet.Payload{
				StopHookActive:       hookInput.StopHookActive,
				LastAssistantMessage: hookInput.LastAssistantMessage,
			},
		}, nil
	default:
		return hookfacet.Event{}, fmt.Errorf("unsupported hook event: %s", hookInput.HookEventName)
	}
}

func (c *ClaudeCode) Write(w io.Writer, ev hookfacet.Decision) error {
	var decision ClaudeCodeDecision
	if ev.Block {
		decision.Decision = "block"
		decision.Reason = ev.Reason
	}
	return json.NewEncoder(w).Encode(decision)
}
