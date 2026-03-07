package api

import (
	"fmt"
	"net/http"

	"github.com/chronick/gangway/internal/translate"
)

// handleImageList handles GET /images/json.
func (s *Server) handleImageList(w http.ResponseWriter, r *http.Request) {
	result, err := s.runner.Run(r.Context(), translate.ImageListArgs()...)
	if err != nil {
		s.logger.Error("image list failed", "error", err)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to list images: %v", err))
		return
	}

	images, err := translate.ParseImageList(result.Stdout)
	if err != nil {
		s.logger.Error("parse image list failed", "error", err)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to parse image list: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, images)
}

// handleImageCreate handles POST /images/create (pull).
func (s *Server) handleImageCreate(w http.ResponseWriter, r *http.Request) {
	imageName := r.URL.Query().Get("fromImage")
	tag := r.URL.Query().Get("tag")
	if tag != "" && tag != "latest" {
		imageName = imageName + ":" + tag
	}

	if imageName == "" {
		writeError(w, http.StatusBadRequest, "fromImage parameter is required")
		return
	}

	result, err := s.runner.Run(r.Context(), translate.ImagePullArgs(imageName)...)
	if err != nil {
		s.logger.Error("image pull failed", "image", imageName, "error", err)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to pull image: %v", err))
		return
	}

	// Docker returns streaming JSON status updates. We simplify to a single response.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"Pull complete","id":"%s"}`, result.Stdout)
}
