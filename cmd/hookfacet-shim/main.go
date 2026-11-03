package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/ironsweet/hookfacet"
	"github.com/ironsweet/hookfacet/wire"
)

const (
	exitAllow = 0
	exitBlock = 2
)

func main() {
	code := run(os.Stdin, os.Stdout, os.Getenv)
	os.Exit(code)
}

type Harness interface {
	Read(r io.Reader) (hookfacet.Event, error)
	Write(w io.Writer, ev hookfacet.Decision) error
}

func run(r io.Reader, w io.Writer, getenv func(string) string) int {
	socketPath := getenv("HOOKFACET_SOCKET")
	if socketPath == "" {
		slog.Error("HOOKFACET_SOCKET not set")
		return exitAllow
	}

	timeoutValue := getenv("HOOKFACET_TIMEOUT")
	timeout, err := time.ParseDuration(timeoutValue)
	if err != nil {
		slog.Warn("Invalid HOOKFACET_TIMEOUT value", "value", timeoutValue, "error", err)
		timeout = 60 * time.Second
	}

	harnessValue := getenv("HOOKFACET_HARNESS")
	if harnessValue == "" {
		harnessValue = "claude-code"
	}
	var harness Harness
	switch harnessValue {
	case "claude-code":
		harness = &ClaudeCode{}
	default:
		slog.Error("Unknown HOOKFACET_HARNESS value", "value", harnessValue)
		return exitAllow
	}

	ev, err := harness.Read(r)
	if err != nil {
		slog.Error("Failed to read event", "error", err)
		return exitAllow
	}

	buf, err := json.Marshal(ev)
	if err != nil {
		slog.Error("Failed to marshal event", "error", err)
		return exitAllow
	}

	conn, err := net.DialTimeout("unix", socketPath, timeout)
	if err != nil {
		slog.Error("Failed to connect to socket", "error", err)
		return exitAllow
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(timeout))
	err = wire.Encode(conn, buf)
	if err != nil {
		slog.Error("Failed to send event", "error", err)
		return exitAllow
	}

	respBuf, err := wire.Decode(conn)
	if err != nil {
		slog.Error("Failed to receive response", "error", err)
		return exitAllow
	}

	var decision hookfacet.Decision
	err = json.Unmarshal(respBuf, &decision)
	if err != nil {
		slog.Error("Failed to unmarshal response", "error", err)
		return exitAllow
	}

	if decision.Block {
		err = harness.Write(w, decision)
		if err != nil {
			slog.Error("Failed to write decision", "error", err)
			return exitAllow
		}
		return exitBlock
	}
	return exitAllow
}

type ClaudeCode struct {
}

func (c *ClaudeCode) Read(r io.Reader) (hookfacet.Event, error) {
	panic("niy")
}

func (c *ClaudeCode) Write(w io.Writer, ev hookfacet.Decision) error {
	panic("niy")
}
