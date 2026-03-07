package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/chronick/gangway/internal/translate"
)

// handleNetworkList handles GET /networks.
func (s *Server) handleNetworkList(w http.ResponseWriter, r *http.Request) {
	result, err := s.runner.Run(r.Context(), translate.NetworkListArgs()...)
	if err != nil {
		s.logger.Error("network list failed", "error", err)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to list networks: %v", err))
		return
	}

	networks, err := translate.ParseNetworkList(result.Stdout)
	if err != nil {
		s.logger.Error("parse network list failed", "error", err)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to parse network list: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, networks)
}

// handleNetworkCreate handles POST /networks/create.
func (s *Server) handleNetworkCreate(w http.ResponseWriter, r *http.Request) {
	var req translate.NetworkCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "network name is required")
		return
	}

	result, err := s.runner.Run(r.Context(), translate.NetworkCreateArgs(req)...)
	if err != nil {
		s.logger.Error("network create failed", "name", req.Name, "error", err)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create network: %v", err))
		return
	}

	// Try to extract network ID from output.
	netID := strings.TrimSpace(result.Stdout)
	if netID == "" {
		netID = generateID()
	}

	writeJSON(w, http.StatusCreated, translate.NetworkCreateResponse{
		ID:      netID,
		Warning: "",
	})
}

// handleNetworkDelete handles DELETE /networks/{id}.
func (s *Server) handleNetworkDelete(w http.ResponseWriter, r *http.Request) {
	id := extractNetworkID(r.URL.Path)

	_, err := s.runner.Run(r.Context(), translate.NetworkDeleteArgs(id)...)
	if err != nil {
		s.logger.Error("network delete failed", "id", id, "error", err)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to delete network: %v", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// extractNetworkID extracts the network ID from a URL path like /v1.41/networks/{id}.
func extractNetworkID(path string) string {
	path = strings.TrimSuffix(path, "/")
	idx := strings.Index(path, "networks/")
	if idx < 0 {
		return ""
	}
	remainder := path[idx+len("networks/"):]
	if slashIdx := strings.Index(remainder, "/"); slashIdx >= 0 {
		return remainder[:slashIdx]
	}
	return remainder
}
