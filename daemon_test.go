package hookfacet

import (
	"context"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/ironsweet/hookfacet/wire"
	"github.com/stretchr/testify/assert"
)

const sampleEventJSON = `{"session_id":"abc123","kind":"turn_end","payload":{}}`

func TestOneCompleteConnection(t *testing.T) {
	server, client := net.Pipe()
	ctx := t.Context()
	runner := NewRunner(&stubHandler{
		mode:        ModeDecide,
		failureMode: FailOpen,
		timeout:     30 * time.Second,
		onEvent: func(ctx context.Context, ev Event, st *State) (Decision, error) {
			assert.Equal(t, "abc123", ev.SessionID)
			assert.Equal(t, KindTurnEnd, ev.Kind)
			return Block("not satisfied"), nil
		},
	})

	connErr := make(chan error, 1)
	go func() {
		connErr <- handleConn(ctx, server, runner)
	}()

	assert.NoError(t, wire.Encode(client, []byte(sampleEventJSON)))
	buf, err := wire.Decode(client)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"block":true,"reason":"not satisfied"}`, string(buf))
	assert.NoError(t, <-connErr)
}

func TestErrorHandling(t *testing.T) {
	server, client := net.Pipe()
	ctx := t.Context()
	runner := NewRunner(&stubHandler{
		mode:        ModeDecide,
		failureMode: FailOpen,
		timeout:     30 * time.Second,
		onEvent: func(ctx context.Context, ev Event, st *State) (Decision, error) {
			return Allow(), assert.AnError
		},
	})

	connErr := make(chan error, 1)
	go func() {
		connErr <- handleConn(ctx, server, runner)
	}()

	assert.NoError(t, wire.Encode(client, []byte(sampleEventJSON)))
	buf, err := wire.Decode(client)
	assert.NoError(t, err)
	assert.JSONEq(t, `{}`, string(buf))
	assert.NoError(t, <-connErr)
}

func TestContextCancel(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "test_cancel.sock")
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	serverErr := make(chan error, 1)
	start := time.Now()
	go func() {
		serverErr <- Serve(ctx, socketPath, NewRunner(&stubHandler{
			mode:        ModeDecide,
			failureMode: FailOpen,
			timeout:     30 * time.Second,
			onEvent: func(ctx context.Context, ev Event, st *State) (Decision, error) {
				return Allow(), nil
			},
		}))
	}()

	var conn net.Conn
	var err error
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		conn, err = net.Dial("unix", socketPath)
		if err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if assert.NoError(t, err) {
		_ = conn.Close()
	}

	cancel()
	select {
	case err = <-serverErr:
		assert.NoError(t, err)
		assert.Less(t, time.Since(start), time.Second, "Serve should return quickly on context cancel")
	case <-time.After(2 * time.Second):
		t.Fatal("Serve did not return after context cancel")
	}
}
