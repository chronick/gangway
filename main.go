package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chronick/gangway/internal/api"
	gexec "github.com/chronick/gangway/internal/exec"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Parse flags.
	socketPath := flag.String("socket", envOrDefault("GANGWAY_SOCKET", "/tmp/gangway.sock"), "unix socket path")
	containerBin := flag.String("container-bin", envOrDefault("GANGWAY_CONTAINER_BIN", "container"), "Apple container CLI path")
	logLevel := flag.String("log-level", envOrDefault("GANGWAY_LOG_LEVEL", "info"), "log level (debug, info, warn, error)")
	flag.Parse()

	// Set up structured logging.
	level := parseLogLevel(*logLevel)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	}))

	logger.Info("gangway starting",
		"socket", *socketPath,
		"container_bin", *containerBin,
		"log_level", *logLevel,
	)

	// Create the exec runner.
	runner := gexec.NewCLIRunner(*containerBin, logger)

	// Create the API server.
	server := api.NewServer(runner, logger)

	// Remove existing socket file if present.
	if err := os.Remove(*socketPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove existing socket: %w", err)
	}

	// Listen on unix socket.
	listener, err := net.Listen("unix", *socketPath)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", *socketPath, err)
	}
	defer listener.Close()

	// Make socket world-readable so containers/agents can connect.
	if err := os.Chmod(*socketPath, 0666); err != nil {
		logger.Warn("failed to chmod socket", "error", err)
	}

	httpServer := &http.Server{
		Handler:      server,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown on SIGTERM/SIGINT.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// Start serving in a goroutine.
	errCh := make(chan error, 1)
	go func() {
		logger.Info("listening", "socket", *socketPath)
		if err := httpServer.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	// Wait for shutdown signal or server error.
	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}
	}

	// Graceful shutdown with timeout.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	logger.Info("gangway stopped")
	return nil
}

// envOrDefault returns the value of the environment variable named key,
// or defaultVal if the variable is not set.
func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// parseLogLevel converts a string log level to slog.Level.
func parseLogLevel(s string) slog.Level {
	switch s {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
