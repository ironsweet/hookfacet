package hookfacet

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"sync"

	"github.com/ironsweet/hookfacet/wire"
)

func Serve(ctx context.Context, socketPath string, runner *Runner, logger *slog.Logger) error {
	if logger == nil {
		logger = slog.Default()
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel() // defensive

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("listen on socket %s: %w", socketPath, err)
	}
	defer listener.Close()

	go func() {
		<-ctx.Done()
		logger.Info("Context canceled, closing listener")
		listener.Close()
	}()

	var wg sync.WaitGroup
	defer wg.Wait() // Wait for all connection handlers to finish

	for {
		var conn net.Conn
		conn, err = listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			logger.Error("Failed to accept connection", "error", err)
			continue
		}
		wg.Go(func() {
			defer conn.Close()
			handleConn(ctx, conn, runner, logger)
		})
	}

}

func handleConn(ctx context.Context, conn net.Conn, runner *Runner, logger *slog.Logger) {
	var ev Event
	buf, err := wire.Decode(conn)
	if err != nil {
		logger.Error("Failed to decode event", "error", err)
		return
	}

	err = json.Unmarshal(buf, &ev)
	if err != nil {
		logger.Error("Failed to unmarshal event", "error", err)
		return
	}

	decision, err := runner.Handle(ctx, ev)
	if err != nil {
		logger.Error("Failed to handle event", "error", err)
		decision = Allow() // Default to allow on error
	}

	respBuf, err := json.Marshal(decision)
	if err != nil {
		logger.Error("Failed to marshal decision", "error", err)
		return
	}

	err = wire.Encode(conn, respBuf)
	if err != nil {
		logger.Error("Failed to encode decision", "error", err)
		return
	}
}
