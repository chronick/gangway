package api

import (
	"encoding/json"
	"net/http"
	"os"
	"runtime"

	"github.com/chronick/gangway/internal/translate"
)

const (
	gangwayVersion = "0.1.0"
	apiVersion     = "1.41"
	minAPIVersion  = "1.24"
)

// handlePing responds to GET /_ping and HEAD /_ping.
func (s *Server) handlePing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("API-Version", apiVersion)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodHead {
		w.Write([]byte("OK"))
	}
}

// handleVersion responds to GET /version.
func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()
	_ = hostname

	resp := translate.VersionResponse{
		Version:       gangwayVersion,
		APIVersion:    apiVersion,
		MinAPIVersion: minAPIVersion,
		GitCommit:     "gangway",
		GoVersion:     runtime.Version(),
		Os:            runtime.GOOS,
		Arch:          runtime.GOARCH,
		KernelVersion: "",
	}

	writeJSON(w, http.StatusOK, resp)
}

// handleInfo responds to GET /info.
func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()

	// Get container counts by listing containers.
	listResult, err := s.runner.Run(r.Context(), translate.ListArgs(true)...)
	containers := 0
	running := 0
	stopped := 0
	if err == nil {
		entries, parseErr := translate.ParseContainerList(listResult.Stdout)
		if parseErr == nil {
			containers = len(entries)
			for _, e := range entries {
				switch e.State {
				case "running":
					running++
				default:
					stopped++
				}
			}
		}
	}

	// Get image count.
	imgResult, err := s.runner.Run(r.Context(), translate.ImageListArgs()...)
	images := 0
	if err == nil {
		imgs, parseErr := translate.ParseImageList(imgResult.Stdout)
		if parseErr == nil {
			images = len(imgs)
		}
	}

	resp := translate.InfoResponse{
		ID:                hostname,
		Containers:        containers,
		ContainersRunning: running,
		ContainersPaused:  0,
		ContainersStopped: stopped,
		Images:            images,
		Driver:            "apple-containerization",
		OperatingSystem:   "macOS",
		OSType:            runtime.GOOS,
		Architecture:      runtime.GOARCH,
		Name:              hostname,
		ServerVersion:     gangwayVersion,
	}

	writeJSON(w, http.StatusOK, resp)
}

// writeJSON marshals v as JSON and writes it to w.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError writes a Docker-style JSON error response.
func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"message": msg})
}
