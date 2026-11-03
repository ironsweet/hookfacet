package main

import (
	"bytes"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/ironsweet/hookfacet/wire"
	"github.com/stretchr/testify/assert"
)

const inputData = `{
  "session_id": "abc123",
  "transcript_path": "~/.claude/projects/.../00893aaf-19fa-41d2-8238-13269b9b3ca0.jsonl",
  "cwd": "/Users/...",
  "permission_mode": "default",
  "hook_event_name": "Stop",
  "stop_hook_active": true,
  "last_assistant_message": "I've completed the refactoring. Here's a summary...",
  "background_tasks": [
    {
      "id": "task-001",
      "type": "shell",
      "status": "running",
      "description": "tail logs",
      "command": "tail -f /var/log/syslog"
    }
  ],
  "session_crons": [
    {
      "id": "cron-001",
      "schedule": "0 9 * * 1-5",
      "recurring": true,
      "prompt": "check the build"
    }
  ]
}`

const normalizedInputData = `{
  "session_id": "abc123",
  "kind": "turn_end",
  "payload": {
    "stop_hook_active":true,
    "last_assistant_message": "I've completed the refactoring. Here's a summary..."
  }
}`

func TestDaemonReturnBlock(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "test1.sock")
	listener, err := net.Listen("unix", socketPath)
	assert.NoError(t, err)
	defer listener.Close()

	daemonErr := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			daemonErr <- err
			return
		}
		defer conn.Close()

		evBuf, err := wire.Decode(conn)
		if err != nil {
			daemonErr <- err
			return
		}
		assert.JSONEq(t, normalizedInputData, string(evBuf))

		err = wire.Encode(conn, []byte(`{"block":true,"reason":"intent not satisfied"}`))
		if err != nil {
			daemonErr <- err
			return
		}
		daemonErr <- nil
	}()

	var outBuf bytes.Buffer
	exitCode := run(bytes.NewBufferString(inputData), &outBuf, func(key string) string {
		switch key {
		case "HOOKFACET_SOCKET":
			return socketPath
		case "HOOKFACET_HARNESS":
			return "claude-code"
		default:
			return ""
		}
	})
	assert.NoError(t, <-daemonErr)
	assert.Equal(t, 2, exitCode)
	assert.JSONEq(t, `{"decision":"block","reason":"intent not satisfied"}`, outBuf.String())
}

func TestDaemonNotRunning(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "wrong.sock")
	var inBuf bytes.Buffer
	var outBuf bytes.Buffer
	exitCode := run(&inBuf, &outBuf, func(key string) string {
		switch key {
		case "HOOKFACET_SOCKET":
			return socketPath
		case "HOOKFACET_HARNESS":
			return "claude-code"
		default:
			return ""
		}
	})
	assert.Equal(t, 0, exitCode)
	assert.Equal(t, 0, outBuf.Len())

	exitCode = run(&inBuf, &outBuf, func(key string) string {
		return ""
	})
	assert.Equal(t, 0, exitCode)
	assert.Equal(t, 0, outBuf.Len())
}

func TestDaemonReturnNonBlock(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "test2.sock")
	listener, err := net.Listen("unix", socketPath)
	assert.NoError(t, err)
	defer listener.Close()

	daemonErr := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			daemonErr <- err
			return
		}
		defer conn.Close()

		evBuf, err := wire.Decode(conn)
		if err != nil {
			daemonErr <- err
			return
		}
		assert.JSONEq(t, normalizedInputData, string(evBuf))

		err = wire.Encode(conn, []byte(`{"block":false}`))
		if err != nil {
			daemonErr <- err
			return
		}
		daemonErr <- nil
	}()

	var outBuf bytes.Buffer
	exitCode := run(bytes.NewBufferString(inputData), &outBuf, func(key string) string {
		switch key {
		case "HOOKFACET_SOCKET":
			return socketPath
		case "HOOKFACET_HARNESS":
			return "claude-code"
		default:
			return ""
		}
	})
	assert.NoError(t, <-daemonErr)
	assert.Equal(t, 0, exitCode)
	assert.Equal(t, 0, outBuf.Len())
}

func TestDaemonReturnTimeout(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "test3.sock")

	listener, err := net.Listen("unix", socketPath)
	assert.NoError(t, err)
	defer listener.Close()

	daemonErr := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			daemonErr <- err
			return
		}
		defer conn.Close()

		evBuf, err := wire.Decode(conn)
		if err != nil {
			daemonErr <- err
			return
		}
		assert.JSONEq(t, normalizedInputData, string(evBuf))

		_ = conn.SetDeadline(time.Now().Add(time.Second))
		_, _ = conn.Read(make([]byte, 1)) // simulate a read timeout
		daemonErr <- nil
	}()

	var outBuf bytes.Buffer
	start := time.Now()
	exitCode := run(bytes.NewBufferString(inputData), &outBuf, func(key string) string {
		switch key {
		case "HOOKFACET_SOCKET":
			return socketPath
		case "HOOKFACET_HARNESS":
			return "claude-code"
		case "HOOKFACET_TIMEOUT":
			return "100ms"
		default:
			return ""
		}
	})
	elapsed := time.Since(start)
	assert.NoError(t, <-daemonErr)
	assert.Equal(t, 0, exitCode)
	assert.Equal(t, 0, outBuf.Len())
	assert.GreaterOrEqual(t, elapsed, 100*time.Millisecond)
	assert.LessOrEqual(t, elapsed, 500*time.Millisecond)
}
