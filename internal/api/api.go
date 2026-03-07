package api

import (
	"log/slog"
	"net/http"
	"strings"
	"sync"

	gexec "github.com/chronick/gangway/internal/exec"
)

// execConfig stores the configuration for a pending exec instance.
type execConfig struct {
	containerID string
	cmd         []string
}

// Server is the Docker API shim HTTP server.
type Server struct {
	runner      gexec.Runner
	logger      *slog.Logger
	mux         *http.ServeMux
	mu          sync.Mutex
	execConfigs map[string]execConfig
}

// NewServer creates a new API server with the given runner and logger.
func NewServer(runner gexec.Runner, logger *slog.Logger) *Server {
	s := &Server{
		runner:      runner,
		logger:      logger,
		mux:         http.NewServeMux(),
		execConfigs: make(map[string]execConfig),
	}
	s.registerRoutes()
	return s
}

// ServeHTTP implements http.Handler. It strips the API version prefix
// (e.g., /v1.41/) before routing to allow versioned and unversioned requests.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Strip version prefix: /v1.41/containers/json -> /containers/json
	path := r.URL.Path
	if len(path) > 2 && path[0] == '/' && path[1] == 'v' {
		if idx := strings.Index(path[2:], "/"); idx >= 0 {
			// Check if the prefix looks like a version number.
			prefix := path[2 : 2+idx]
			if isVersionPrefix(prefix) {
				r.URL.Path = path[2+idx:]
				path = r.URL.Path
			}
		}
	}

	s.logger.Debug("request", "method", r.Method, "path", path)
	s.mux.ServeHTTP(w, r)
}

// registerRoutes sets up all Docker API routes.
func (s *Server) registerRoutes() {
	// System
	s.mux.HandleFunc("/_ping", s.handlePing)
	s.mux.HandleFunc("/version", s.methodGuard(http.MethodGet, s.handleVersion))
	s.mux.HandleFunc("/info", s.methodGuard(http.MethodGet, s.handleInfo))

	// Containers
	s.mux.HandleFunc("/containers/json", s.methodGuard(http.MethodGet, s.handleContainerList))
	s.mux.HandleFunc("/containers/create", s.methodGuard(http.MethodPost, s.handleContainerCreate))

	// Container operations use a pattern with the ID in the path.
	// Since net/http.ServeMux doesn't support path parameters, we use a
	// catch-all handler for /containers/ and dispatch internally.
	s.mux.HandleFunc("/containers/", s.routeContainer)

	// Images
	s.mux.HandleFunc("/images/json", s.methodGuard(http.MethodGet, s.handleImageList))
	s.mux.HandleFunc("/images/create", s.methodGuard(http.MethodPost, s.handleImageCreate))

	// Networks
	s.mux.HandleFunc("/networks", s.methodGuard(http.MethodGet, s.handleNetworkList))
	s.mux.HandleFunc("/networks/create", s.methodGuard(http.MethodPost, s.handleNetworkCreate))
	s.mux.HandleFunc("/networks/", s.routeNetwork)

	// Exec
	s.mux.HandleFunc("/exec/", s.routeExec)
}

// routeContainer dispatches /containers/{id}/* requests to the correct handler.
func (s *Server) routeContainer(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch {
	case strings.HasSuffix(path, "/json"):
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		s.handleContainerInspect(w, r)

	case strings.HasSuffix(path, "/start"):
		if r.Method != http.MethodPost {
			methodNotAllowed(w)
			return
		}
		s.handleContainerStart(w, r)

	case strings.HasSuffix(path, "/stop"):
		if r.Method != http.MethodPost {
			methodNotAllowed(w)
			return
		}
		s.handleContainerStop(w, r)

	case strings.HasSuffix(path, "/kill"):
		if r.Method != http.MethodPost {
			methodNotAllowed(w)
			return
		}
		s.handleContainerKill(w, r)

	case strings.HasSuffix(path, "/logs"):
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		s.handleContainerLogs(w, r)

	case strings.HasSuffix(path, "/exec"):
		if r.Method != http.MethodPost {
			methodNotAllowed(w)
			return
		}
		s.handleContainerExecCreate(w, r)

	default:
		// DELETE /containers/{id}
		if r.Method == http.MethodDelete {
			s.handleContainerRemove(w, r)
			return
		}
		notImplemented(w)
	}
}

// routeNetwork dispatches /networks/{id} requests.
func (s *Server) routeNetwork(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Skip /networks/create — already registered directly.
	if strings.HasSuffix(path, "/create") {
		// This shouldn't happen since /networks/create is registered first,
		// but handle it defensively.
		s.handleNetworkCreate(w, r)
		return
	}

	if r.Method == http.MethodDelete {
		s.handleNetworkDelete(w, r)
		return
	}

	notImplemented(w)
}

// routeExec dispatches /exec/{id}/* requests.
func (s *Server) routeExec(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/start") {
		if r.Method != http.MethodPost {
			methodNotAllowed(w)
			return
		}
		s.handleExecStart(w, r)
		return
	}
	notImplemented(w)
}

// methodGuard wraps a handler to only allow the specified HTTP method.
func (s *Server) methodGuard(method string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			methodNotAllowed(w)
			return
		}
		handler(w, r)
	}
}

// notImplemented returns a 501 response for unsupported endpoints.
func notImplemented(w http.ResponseWriter) {
	writeError(w, http.StatusNotImplemented, "endpoint not implemented by gangway")
}

// methodNotAllowed returns a 405 response.
func methodNotAllowed(w http.ResponseWriter) {
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}

// isVersionPrefix checks if a string looks like a Docker API version (e.g., "1.41").
func isVersionPrefix(s string) bool {
	// Simple check: contains a dot and only digits/dots.
	hasDot := false
	for _, c := range s {
		if c == '.' {
			hasDot = true
		} else if c < '0' || c > '9' {
			return false
		}
	}
	return hasDot
}
