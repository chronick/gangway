package exec

import (
	"context"
	"errors"
	"testing"
)

func TestMockRunner_RecordsCalls(t *testing.T) {
	mock := NewMockRunner()
	mock.DefaultResult = MockResult{
		Result: Result{Stdout: "ok\n", ExitCode: 0},
	}

	ctx := context.Background()
	_, _ = mock.Run(ctx, "list", "--format", "json")
	_, _ = mock.Run(ctx, "inspect", "abc123")

	if len(mock.Calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(mock.Calls))
	}

	if mock.Calls[0].Args[0] != "list" {
		t.Errorf("expected first call arg 'list', got %q", mock.Calls[0].Args[0])
	}
	if mock.Calls[1].Args[0] != "inspect" {
		t.Errorf("expected second call arg 'inspect', got %q", mock.Calls[1].Args[0])
	}
}

func TestMockRunner_ReturnsQueuedResults(t *testing.T) {
	mock := NewMockRunner()
	mock.PushResult(`[{"id":"a"}]`, 0, nil)
	mock.PushResult("", 1, errors.New("not found"))

	ctx := context.Background()

	r1, err1 := mock.Run(ctx, "list")
	if err1 != nil {
		t.Errorf("expected no error on first call, got %v", err1)
	}
	if r1.Stdout != `[{"id":"a"}]` {
		t.Errorf("unexpected stdout: %q", r1.Stdout)
	}

	r2, err2 := mock.Run(ctx, "inspect", "missing")
	if err2 == nil {
		t.Error("expected error on second call")
	}
	if r2.ExitCode != 1 {
		t.Errorf("expected exit code 1, got %d", r2.ExitCode)
	}
}

func TestMockRunner_FallsBackToDefault(t *testing.T) {
	mock := NewMockRunner()
	mock.DefaultResult = MockResult{
		Result: Result{Stdout: "default", ExitCode: 0},
	}

	ctx := context.Background()
	r, err := mock.Run(ctx, "anything")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if r.Stdout != "default" {
		t.Errorf("expected default stdout, got %q", r.Stdout)
	}
}

func TestMockRunner_EmptyQueue_NoDefault(t *testing.T) {
	mock := NewMockRunner()

	ctx := context.Background()
	r, err := mock.Run(ctx, "anything")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if r.Stdout != "" {
		t.Errorf("expected empty stdout, got %q", r.Stdout)
	}
	if r.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", r.ExitCode)
	}
}
