# CLAUDE.md

## Project

**gangway** -- Docker API shim for Apple containers.

Go 1.22+, zero external dependencies. Translates Docker Engine API calls
to Apple `container` CLI commands over a unix socket.

## Commands

```bash
go build -o gangway .        # build
go test ./...                # test all
```

## Layout

```
main.go                      # entry point, socket listener
internal/
  api/api.go                 # HTTP mux, Docker API route dispatch
  api/containers.go          # /containers/* handlers
  api/images.go              # /images/* handlers
  api/networks.go            # /networks/* handlers
  api/system.go              # /_ping, /version, /info
  translate/translate.go     # Docker API types -> container CLI args
  translate/parse.go         # container CLI output -> Docker API types
  exec/exec.go               # os/exec wrapper for container CLI
```

## Key Design

- Stateless: all state lives in containermanagerd, gangway just translates
- Docker API subset: only endpoints needed by OpenClaw + dockerode
- Unsupported endpoints return 501, not silent failures
- Container names map 1:1 (Docker name = Apple container name)
- Uses os/exec to call `container` CLI, not Mach IPC directly
- Structured logging via log/slog

## Testing Strategy

- Unit tests: mock exec runner, verify CLI args generated from API calls
- Integration tests: require `container` CLI, create/inspect/delete real containers
- Conformance tests: run dockerode test suite against gangway socket

## Beads Workflow Integration

<!-- br-agent-instructions-v1 -->
See skiff CLAUDE.md for beads workflow details.
<!-- end-br-agent-instructions -->
