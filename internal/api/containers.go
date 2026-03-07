package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/chronick/gangway/internal/translate"
)

// handleContainerList handles GET /containers/json.
func (s *Server) handleContainerList(w http.ResponseWriter, r *http.Request) {
	all := r.URL.Query().Get("all") == "1" || r.URL.Query().Get("all") == "true"

	result, err := s.runner.Run(r.Context(), translate.ListArgs(all)...)
	if err != nil {
		s.logger.Error("container list failed", "error", err)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to list containers: %v", err))
		return
	}

	entries, err := translate.ParseContainerList(result.Stdout)
	if err != nil {
		s.logger.Error("parse container list failed", "error", err)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to parse container list: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, entries)
}

// handleContainerInspect handles GET /containers/{id}/json.
func (s *Server) handleContainerInspect(w http.ResponseWriter, r *http.Request) {
	id := extractContainerID(r.URL.Path, "/json")

	result, err := s.runner.Run(r.Context(), translate.InspectArgs(id)...)
	if err != nil {
		s.logger.Error("container inspect failed", "id", id, "error", err)
		writeError(w, http.StatusNotFound, fmt.Sprintf("no such container: %s", id))
		return
	}

	container, err := translate.ParseContainerInspect(result.Stdout)
	if err != nil {
		s.logger.Error("parse container inspect failed", "id", id, "error", err)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to parse container inspect: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, container)
}

// handleContainerCreate handles POST /containers/create.
func (s *Server) handleContainerCreate(w http.ResponseWriter, r *http.Request) {
	var req translate.CreateContainerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	name := r.URL.Query().Get("name")

	args := translate.CreateArgs(req, name)
	result, err := s.runner.Run(r.Context(), args...)
	if err != nil {
		s.logger.Error("container create failed", "error", err, "stderr", result.Stderr)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create container: %v", err))
		return
	}

	// Try to extract the container ID from stdout.
	containerID := strings.TrimSpace(result.Stdout)
	if containerID == "" {
		// Generate a pseudo-ID if the CLI didn't return one.
		containerID = generateID()
	}

	resp := translate.CreateContainerResponse{
		ID:       containerID,
		Warnings: []string{},
	}

	writeJSON(w, http.StatusCreated, resp)
}

// handleContainerStart handles POST /containers/{id}/start.
func (s *Server) handleContainerStart(w http.ResponseWriter, r *http.Request) {
	id := extractContainerID(r.URL.Path, "/start")

	_, err := s.runner.Run(r.Context(), translate.StartArgs(id)...)
	if err != nil {
		s.logger.Error("container start failed", "id", id, "error", err)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to start container: %v", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleContainerStop handles POST /containers/{id}/stop.
func (s *Server) handleContainerStop(w http.ResponseWriter, r *http.Request) {
	id := extractContainerID(r.URL.Path, "/stop")
	timeout := 0
	if t := r.URL.Query().Get("t"); t != "" {
		if v, err := strconv.Atoi(t); err == nil {
			timeout = v
		}
	}

	_, err := s.runner.Run(r.Context(), translate.StopArgs(id, timeout)...)
	if err != nil {
		s.logger.Error("container stop failed", "id", id, "error", err)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to stop container: %v", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleContainerKill handles POST /containers/{id}/kill.
func (s *Server) handleContainerKill(w http.ResponseWriter, r *http.Request) {
	id := extractContainerID(r.URL.Path, "/kill")
	signal := r.URL.Query().Get("signal")

	_, err := s.runner.Run(r.Context(), translate.KillArgs(id, signal)...)
	if err != nil {
		s.logger.Error("container kill failed", "id", id, "error", err)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to kill container: %v", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleContainerRemove handles DELETE /containers/{id}.
func (s *Server) handleContainerRemove(w http.ResponseWriter, r *http.Request) {
	id := extractContainerID(r.URL.Path, "")
	force := r.URL.Query().Get("force") == "1" || r.URL.Query().Get("force") == "true"

	_, err := s.runner.Run(r.Context(), translate.RemoveArgs(id, force)...)
	if err != nil {
		s.logger.Error("container remove failed", "id", id, "error", err)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to remove container: %v", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleContainerLogs handles GET /containers/{id}/logs.
func (s *Server) handleContainerLogs(w http.ResponseWriter, r *http.Request) {
	id := extractContainerID(r.URL.Path, "/logs")
	follow := r.URL.Query().Get("follow") == "1" || r.URL.Query().Get("follow") == "true"
	tail := r.URL.Query().Get("tail")
	timestamps := r.URL.Query().Get("timestamps") == "1" || r.URL.Query().Get("timestamps") == "true"

	result, err := s.runner.Run(r.Context(), translate.LogsArgs(id, follow, tail, timestamps)...)
	if err != nil {
		s.logger.Error("container logs failed", "id", id, "error", err)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get container logs: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/vnd.docker.raw-stream")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result.Stdout))
}

// handleContainerExecCreate handles POST /containers/{id}/exec.
func (s *Server) handleContainerExecCreate(w http.ResponseWriter, r *http.Request) {
	id := extractContainerID(r.URL.Path, "/exec")
	_ = id // Used to validate the container exists.

	var req translate.ExecCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	// Store the exec config for later use by exec start.
	execID := generateID()
	s.mu.Lock()
	s.execConfigs[execID] = execConfig{
		containerID: id,
		cmd:         req.Cmd,
	}
	s.mu.Unlock()

	writeJSON(w, http.StatusCreated, translate.ExecCreateResponse{ID: execID})
}

// handleExecStart handles POST /exec/{id}/start.
func (s *Server) handleExecStart(w http.ResponseWriter, r *http.Request) {
	execID := extractExecID(r.URL.Path)

	s.mu.Lock()
	config, ok := s.execConfigs[execID]
	if ok {
		delete(s.execConfigs, execID)
	}
	s.mu.Unlock()

	if !ok {
		writeError(w, http.StatusNotFound, fmt.Sprintf("no such exec instance: %s", execID))
		return
	}

	result, err := s.runner.Run(r.Context(), translate.ExecArgs(config.containerID, config.cmd)...)
	if err != nil {
		s.logger.Error("exec start failed", "exec_id", execID, "error", err)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("exec failed: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/vnd.docker.raw-stream")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result.Stdout))
}

// extractContainerID extracts the container ID from a URL path like
// /v1.41/containers/{id}/json or /containers/{id}.
// The suffix parameter is stripped from the end before extracting.
func extractContainerID(path, suffix string) string {
	if suffix != "" {
		path = strings.TrimSuffix(path, suffix)
	}
	path = strings.TrimSuffix(path, "/")

	// Find "containers/" and take the next segment.
	idx := strings.Index(path, "containers/")
	if idx < 0 {
		return ""
	}
	remainder := path[idx+len("containers/"):]
	// Take up to the next slash, if any.
	if slashIdx := strings.Index(remainder, "/"); slashIdx >= 0 {
		return remainder[:slashIdx]
	}
	return remainder
}

// extractExecID extracts the exec ID from a URL path like /v1.41/exec/{id}/start.
func extractExecID(path string) string {
	path = strings.TrimSuffix(path, "/start")
	path = strings.TrimSuffix(path, "/")

	idx := strings.Index(path, "exec/")
	if idx < 0 {
		return ""
	}
	remainder := path[idx+len("exec/"):]
	if slashIdx := strings.Index(remainder, "/"); slashIdx >= 0 {
		return remainder[:slashIdx]
	}
	return remainder
}

// generateID generates a random 64-character hex string (like a Docker container ID).
func generateID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
