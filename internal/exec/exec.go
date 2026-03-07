package exec

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os/exec"
)

// Result holds the output of a CLI command.
type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Runner executes container CLI commands.
type Runner interface {
	// Run executes the container CLI with the given arguments.
	Run(ctx context.Context, args ...string) (Result, error)
}

// CLIRunner is the real implementation that calls the container CLI via os/exec.
type CLIRunner struct {
	// BinPath is the path to the container CLI binary.
	BinPath string
	// Logger for structured logging.
	Logger *slog.Logger
}

// NewCLIRunner creates a new CLIRunner with the given binary path.
func NewCLIRunner(binPath string, logger *slog.Logger) *CLIRunner {
	return &CLIRunner{
		BinPath: binPath,
		Logger:  logger,
	}
}

// Run executes the container CLI with the given arguments.
func (r *CLIRunner) Run(ctx context.Context, args ...string) (Result, error) {
	r.Logger.Debug("exec", "bin", r.BinPath, "args", args)

	cmd := exec.CommandContext(ctx, r.BinPath, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := Result{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			r.Logger.Warn("exec failed",
				"bin", r.BinPath,
				"args", args,
				"exit_code", result.ExitCode,
				"stderr", result.Stderr,
			)
			return result, fmt.Errorf("command exited with code %d: %s", result.ExitCode, result.Stderr)
		}
		return result, fmt.Errorf("command execution failed: %w", err)
	}

	r.Logger.Debug("exec ok", "bin", r.BinPath, "args", args, "stdout_len", len(result.Stdout))
	return result, nil
}

// MockRunner is a test implementation of Runner.
type MockRunner struct {
	// Calls records all invocations for assertion.
	Calls []MockCall
	// Results is a queue of results to return. Shifted on each call.
	Results []MockResult
	// DefaultResult is returned when Results is empty.
	DefaultResult MockResult
}

// MockCall records a single invocation.
type MockCall struct {
	Args []string
}

// MockResult is a predefined result for MockRunner.
type MockResult struct {
	Result Result
	Err    error
}

// NewMockRunner creates a new MockRunner.
func NewMockRunner() *MockRunner {
	return &MockRunner{}
}

// PushResult enqueues a result to be returned on the next call.
func (m *MockRunner) PushResult(stdout string, exitCode int, err error) {
	m.Results = append(m.Results, MockResult{
		Result: Result{Stdout: stdout, ExitCode: exitCode},
		Err:    err,
	})
}

// Run records the call and returns the next queued result.
func (m *MockRunner) Run(_ context.Context, args ...string) (Result, error) {
	m.Calls = append(m.Calls, MockCall{Args: args})

	if len(m.Results) > 0 {
		r := m.Results[0]
		m.Results = m.Results[1:]
		return r.Result, r.Err
	}

	return m.DefaultResult.Result, m.DefaultResult.Err
}
