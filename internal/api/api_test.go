package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	gexec "github.com/chronick/gangway/internal/exec"
	"github.com/chronick/gangway/internal/translate"
)

func newTestServer() (*Server, *gexec.MockRunner) {
	mock := gexec.NewMockRunner()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := NewServer(mock, logger)
	return srv, mock
}

func TestPing(t *testing.T) {
	srv, _ := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/_ping", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "OK" {
		t.Errorf("expected 'OK', got %q", w.Body.String())
	}
	if w.Header().Get("API-Version") != apiVersion {
		t.Errorf("expected API-Version header %q", apiVersion)
	}
}

func TestPingHead(t *testing.T) {
	srv, _ := newTestServer()
	req := httptest.NewRequest(http.MethodHead, "/_ping", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.Len() != 0 {
		t.Errorf("expected empty body for HEAD, got %q", w.Body.String())
	}
}

func TestVersion(t *testing.T) {
	srv, _ := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp translate.VersionResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Version != gangwayVersion {
		t.Errorf("expected version %q, got %q", gangwayVersion, resp.Version)
	}
	if resp.APIVersion != apiVersion {
		t.Errorf("expected API version %q, got %q", apiVersion, resp.APIVersion)
	}
}

func TestVersionedRoute(t *testing.T) {
	srv, _ := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/v1.41/_ping", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for versioned _ping, got %d", w.Code)
	}
	if w.Body.String() != "OK" {
		t.Errorf("expected 'OK', got %q", w.Body.String())
	}
}

func TestContainerList(t *testing.T) {
	srv, mock := newTestServer()
	mock.PushResult(`[{"id":"abc123","name":"web","image":"nginx","status":"running","createdAt":"2025-01-01T00:00:00Z","command":"nginx"}]`, 0, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1.41/containers/json", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var entries []translate.ContainerListEntry
	if err := json.NewDecoder(w.Body).Decode(&entries); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].ID != "abc123" {
		t.Errorf("expected ID abc123, got %q", entries[0].ID)
	}
}

func TestContainerInspect(t *testing.T) {
	srv, mock := newTestServer()
	mock.PushResult(`{"id":"abc123","name":"web","image":"nginx","status":"running","pid":1234,"exitCode":0,"createdAt":"2025-01-01T00:00:00Z","startedAt":"","stoppedAt":"","command":"nginx","env":[]}`, 0, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1.41/containers/abc123/json", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var container translate.ContainerJSON
	if err := json.NewDecoder(w.Body).Decode(&container); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if container.ID != "abc123" {
		t.Errorf("expected ID abc123, got %q", container.ID)
	}
}

func TestContainerCreate(t *testing.T) {
	srv, mock := newTestServer()
	mock.PushResult("container-id-xyz\n", 0, nil)

	body := `{"Image":"alpine:latest","Cmd":["echo","hello"]}`
	req := httptest.NewRequest(http.MethodPost, "/v1.41/containers/create?name=mytest", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}

	var resp translate.CreateContainerResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp.ID != "container-id-xyz" {
		t.Errorf("expected ID container-id-xyz, got %q", resp.ID)
	}

	// Verify the runner was called with correct args.
	if len(mock.Calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mock.Calls))
	}
	args := mock.Calls[0].Args
	if args[0] != "run" {
		t.Errorf("expected 'run' as first arg, got %q", args[0])
	}
}

func TestContainerStart(t *testing.T) {
	srv, mock := newTestServer()
	mock.PushResult("", 0, nil)

	req := httptest.NewRequest(http.MethodPost, "/v1.41/containers/abc123/start", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}

	if len(mock.Calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mock.Calls))
	}
	if mock.Calls[0].Args[0] != "start" {
		t.Errorf("expected 'start', got %q", mock.Calls[0].Args[0])
	}
}

func TestContainerStop(t *testing.T) {
	srv, mock := newTestServer()
	mock.PushResult("", 0, nil)

	req := httptest.NewRequest(http.MethodPost, "/v1.41/containers/abc123/stop?t=10", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

func TestContainerKill(t *testing.T) {
	srv, mock := newTestServer()
	mock.PushResult("", 0, nil)

	req := httptest.NewRequest(http.MethodPost, "/v1.41/containers/abc123/kill?signal=SIGKILL", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}

	if mock.Calls[0].Args[0] != "kill" {
		t.Errorf("expected 'kill', got %q", mock.Calls[0].Args[0])
	}
}

func TestContainerRemove(t *testing.T) {
	srv, mock := newTestServer()
	mock.PushResult("", 0, nil)

	req := httptest.NewRequest(http.MethodDelete, "/v1.41/containers/abc123?force=true", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}

	if mock.Calls[0].Args[0] != "delete" {
		t.Errorf("expected 'delete', got %q", mock.Calls[0].Args[0])
	}
}

func TestContainerLogs(t *testing.T) {
	srv, mock := newTestServer()
	mock.PushResult("log line 1\nlog line 2\n", 0, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1.41/containers/abc123/logs?tail=100", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "log line 1\nlog line 2\n" {
		t.Errorf("unexpected body: %q", w.Body.String())
	}
}

func TestImageList(t *testing.T) {
	srv, mock := newTestServer()
	mock.PushResult(`[{"id":"sha256:abc","tags":["alpine:latest"],"digests":[],"createdAt":"2025-01-01T00:00:00Z","size":7340032}]`, 0, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1.41/images/json", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var images []translate.ImageListEntry
	if err := json.NewDecoder(w.Body).Decode(&images); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(images) != 1 {
		t.Fatalf("expected 1 image, got %d", len(images))
	}
}

func TestImagePull(t *testing.T) {
	srv, mock := newTestServer()
	mock.PushResult("pulled\n", 0, nil)

	req := httptest.NewRequest(http.MethodPost, "/v1.41/images/create?fromImage=alpine&tag=latest", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	if mock.Calls[0].Args[0] != "image" || mock.Calls[0].Args[1] != "pull" {
		t.Errorf("expected 'image pull', got %v", mock.Calls[0].Args)
	}
}

func TestNetworkList(t *testing.T) {
	srv, mock := newTestServer()
	mock.PushResult(`[{"id":"net1","name":"bridge","driver":"bridge"}]`, 0, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1.41/networks", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestNetworkCreate(t *testing.T) {
	srv, mock := newTestServer()
	mock.PushResult("net-abc\n", 0, nil)

	body := `{"Name":"mynet","Driver":"bridge"}`
	req := httptest.NewRequest(http.MethodPost, "/v1.41/networks/create", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
}

func TestNetworkDelete(t *testing.T) {
	srv, mock := newTestServer()
	mock.PushResult("", 0, nil)

	req := httptest.NewRequest(http.MethodDelete, "/v1.41/networks/mynet", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}

	if mock.Calls[0].Args[0] != "network" || mock.Calls[0].Args[1] != "delete" {
		t.Errorf("expected 'network delete', got %v", mock.Calls[0].Args)
	}
}

func TestExecCreateAndStart(t *testing.T) {
	srv, mock := newTestServer()

	// Push result for exec start.
	mock.PushResult("hello world\n", 0, nil)

	// Create exec.
	body := `{"Cmd":["echo","hello"],"AttachStdout":true}`
	req := httptest.NewRequest(http.MethodPost, "/v1.41/containers/abc123/exec", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("exec create: expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var execResp translate.ExecCreateResponse
	if err := json.NewDecoder(w.Body).Decode(&execResp); err != nil {
		t.Fatalf("failed to decode exec create response: %v", err)
	}
	if execResp.ID == "" {
		t.Fatal("expected non-empty exec ID")
	}

	// Start exec.
	startBody := `{"Detach":false}`
	req = httptest.NewRequest(http.MethodPost, "/v1.41/exec/"+execResp.ID+"/start", bytes.NewBufferString(startBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("exec start: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if w.Body.String() != "hello world\n" {
		t.Errorf("expected 'hello world\\n', got %q", w.Body.String())
	}
}

func TestMethodNotAllowed(t *testing.T) {
	srv, _ := newTestServer()

	req := httptest.NewRequest(http.MethodPost, "/version", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestNotImplemented(t *testing.T) {
	srv, _ := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/v1.41/containers/abc123/changes", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Errorf("expected 501, got %d", w.Code)
	}
}

func TestExtractContainerID(t *testing.T) {
	tests := []struct {
		path     string
		suffix   string
		expected string
	}{
		{"/containers/abc123/json", "/json", "abc123"},
		{"/v1.41/containers/abc123/json", "/json", "abc123"},
		{"/containers/abc123/start", "/start", "abc123"},
		{"/containers/abc123/stop", "/stop", "abc123"},
		{"/containers/abc123/kill", "/kill", "abc123"},
		{"/containers/abc123/logs", "/logs", "abc123"},
		{"/containers/abc123/exec", "/exec", "abc123"},
		{"/containers/abc123", "", "abc123"},
	}

	for _, tt := range tests {
		got := extractContainerID(tt.path, tt.suffix)
		if got != tt.expected {
			t.Errorf("extractContainerID(%q, %q) = %q, want %q", tt.path, tt.suffix, got, tt.expected)
		}
	}
}

func TestExtractExecID(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/exec/abc123/start", "abc123"},
		{"/v1.41/exec/abc123/start", "abc123"},
	}

	for _, tt := range tests {
		got := extractExecID(tt.path)
		if got != tt.expected {
			t.Errorf("extractExecID(%q) = %q, want %q", tt.path, got, tt.expected)
		}
	}
}

func TestExtractNetworkID(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/networks/mynet", "mynet"},
		{"/v1.41/networks/mynet", "mynet"},
	}

	for _, tt := range tests {
		got := extractNetworkID(tt.path)
		if got != tt.expected {
			t.Errorf("extractNetworkID(%q) = %q, want %q", tt.path, got, tt.expected)
		}
	}
}

func TestIsVersionPrefix(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"1.41", true},
		{"1.24", true},
		{"1.0", true},
		{"abc", false},
		{"1", false},
		{"", false},
	}

	for _, tt := range tests {
		got := isVersionPrefix(tt.input)
		if got != tt.expected {
			t.Errorf("isVersionPrefix(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}
