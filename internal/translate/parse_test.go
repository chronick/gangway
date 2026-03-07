package translate

import (
	"testing"
)

func TestParseContainerList_Empty(t *testing.T) {
	result, err := ParseContainerList("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty list, got %d entries", len(result))
	}
}

func TestParseContainerList_SingleContainer(t *testing.T) {
	input := `[{"id":"abc123","name":"web","image":"nginx:latest","status":"running","createdAt":"2025-01-01T00:00:00Z","command":"nginx -g daemon off;"}]`

	result, err := ParseContainerList(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}

	c := result[0]
	if c.ID != "abc123" {
		t.Errorf("expected ID abc123, got %q", c.ID)
	}
	if len(c.Names) != 1 || c.Names[0] != "/web" {
		t.Errorf("expected Names [/web], got %v", c.Names)
	}
	if c.Image != "nginx:latest" {
		t.Errorf("expected Image nginx:latest, got %q", c.Image)
	}
	if c.State != "running" {
		t.Errorf("expected State running, got %q", c.State)
	}
	if c.Command != "nginx -g daemon off;" {
		t.Errorf("expected Command, got %q", c.Command)
	}
	if c.Created == 0 {
		t.Error("expected non-zero Created timestamp")
	}
}

func TestParseContainerList_MultipleContainers(t *testing.T) {
	input := `[
		{"id":"a1","name":"web","image":"nginx","status":"running","createdAt":"2025-01-01T00:00:00Z","command":""},
		{"id":"b2","name":"db","image":"postgres","status":"stopped","createdAt":"2025-01-02T00:00:00Z","command":""}
	]`

	result, err := ParseContainerList(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
	if result[0].ID != "a1" || result[1].ID != "b2" {
		t.Errorf("unexpected IDs: %q, %q", result[0].ID, result[1].ID)
	}
	if result[1].State != "exited" {
		t.Errorf("expected stopped -> exited, got %q", result[1].State)
	}
}

func TestParseContainerList_SingleObject(t *testing.T) {
	input := `{"id":"abc123","name":"solo","image":"alpine","status":"running","createdAt":"","command":""}`

	result, err := ParseContainerList(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}
	if result[0].ID != "abc123" {
		t.Errorf("expected ID abc123, got %q", result[0].ID)
	}
}

func TestParseContainerInspect(t *testing.T) {
	input := `{"id":"abc123","name":"web","image":"nginx:latest","status":"running","pid":1234,"exitCode":0,"createdAt":"2025-01-01T00:00:00Z","startedAt":"2025-01-01T00:00:01Z","stoppedAt":"","command":"nginx -g daemon off;","env":["FOO=bar"]}`

	result, err := ParseContainerInspect(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "abc123" {
		t.Errorf("expected ID abc123, got %q", result.ID)
	}
	if result.Name != "/web" {
		t.Errorf("expected Name /web, got %q", result.Name)
	}
	if !result.State.Running {
		t.Error("expected State.Running to be true")
	}
	if result.State.Pid != 1234 {
		t.Errorf("expected Pid 1234, got %d", result.State.Pid)
	}
	if result.Config.Image != "nginx:latest" {
		t.Errorf("expected Config.Image nginx:latest, got %q", result.Config.Image)
	}
	if len(result.Config.Env) != 1 || result.Config.Env[0] != "FOO=bar" {
		t.Errorf("expected Env [FOO=bar], got %v", result.Config.Env)
	}
}

func TestParseContainerInspect_Stopped(t *testing.T) {
	input := `{"id":"def456","name":"done","image":"alpine","status":"stopped","pid":0,"exitCode":0,"createdAt":"2025-01-01T00:00:00Z","startedAt":"2025-01-01T00:00:01Z","stoppedAt":"2025-01-01T00:01:00Z","command":"echo hello","env":[]}`

	result, err := ParseContainerInspect(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.State.Running {
		t.Error("expected State.Running to be false")
	}
	if result.State.Status != "exited" {
		t.Errorf("expected Status exited, got %q", result.State.Status)
	}
}

func TestParseContainerInspect_Empty(t *testing.T) {
	_, err := ParseContainerInspect("")
	if err == nil {
		t.Error("expected error for empty input")
	}
}

func TestParseContainerInspect_Array(t *testing.T) {
	input := `[{"id":"abc123","name":"web","image":"nginx","status":"running","pid":0,"exitCode":0,"createdAt":"","startedAt":"","stoppedAt":"","command":"","env":[]}]`

	result, err := ParseContainerInspect(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "abc123" {
		t.Errorf("expected ID abc123, got %q", result.ID)
	}
}

func TestParseImageList_Empty(t *testing.T) {
	result, err := ParseImageList("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty list, got %d", len(result))
	}
}

func TestParseImageList(t *testing.T) {
	input := `[{"id":"sha256:abc","tags":["alpine:latest","alpine:3.19"],"digests":["sha256:def"],"createdAt":"2025-01-01T00:00:00Z","size":7340032}]`

	result, err := ParseImageList(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}

	img := result[0]
	if img.ID != "sha256:abc" {
		t.Errorf("expected ID sha256:abc, got %q", img.ID)
	}
	if len(img.RepoTags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(img.RepoTags))
	}
	if img.Size != 7340032 {
		t.Errorf("expected Size 7340032, got %d", img.Size)
	}
}

func TestParseImageList_NilTags(t *testing.T) {
	input := `[{"id":"sha256:abc","createdAt":"","size":0}]`

	result, err := ParseImageList(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result[0].RepoTags == nil {
		t.Error("expected non-nil RepoTags")
	}
	if result[0].RepoDigests == nil {
		t.Error("expected non-nil RepoDigests")
	}
}

func TestParseNetworkList_Empty(t *testing.T) {
	result, err := ParseNetworkList("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty list, got %d", len(result))
	}
}

func TestParseNetworkList(t *testing.T) {
	input := `[{"id":"net1","name":"bridge","driver":"bridge"},{"id":"net2","name":"custom","driver":"overlay"}]`

	result, err := ParseNetworkList(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
	if result[0].Name != "bridge" {
		t.Errorf("expected Name bridge, got %q", result[0].Name)
	}
	if result[0].Scope != "local" {
		t.Errorf("expected Scope local, got %q", result[0].Scope)
	}
}

func TestNormalizeState(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"running", "running"},
		{"Running", "running"},
		{"stopped", "exited"},
		{"Stopped", "exited"},
		{"exited", "exited"},
		{"created", "created"},
		{"paused", "paused"},
		{"", ""},
	}

	for _, tt := range tests {
		got := normalizeState(tt.input)
		if got != tt.expected {
			t.Errorf("normalizeState(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestParseTimeUnix(t *testing.T) {
	// Valid RFC3339
	ts := parseTimeUnix("2025-01-01T00:00:00Z")
	if ts == 0 {
		t.Error("expected non-zero timestamp for valid RFC3339")
	}

	// Empty string
	ts = parseTimeUnix("")
	if ts != 0 {
		t.Error("expected 0 for empty string")
	}

	// Invalid string
	ts = parseTimeUnix("not-a-date")
	if ts != 0 {
		t.Error("expected 0 for invalid date")
	}
}
