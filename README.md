# gangway

Docker API shim for Apple containers. Exposes a Docker-compatible unix socket that translates Docker Engine API calls to Apple `container` CLI commands.

Any tool that expects a Docker socket (OpenClaw, dockerode, CI pipelines, Testcontainers) works unmodified with Apple containers.

## Install

```bash
go install github.com/chronick/gangway@latest
```

## Usage

```bash
# Start the shim (listens on unix socket)
gangway --socket /tmp/gangway.sock

# Point Docker-expecting tools at it
export DOCKER_HOST=unix:///tmp/gangway.sock
export OPENCLAW_DOCKER_SOCKET=/tmp/gangway.sock

# Verify
curl --unix-socket /tmp/gangway.sock http://localhost/_ping
# OK
```

## What It Translates

| Docker API | Apple Container CLI |
|-----------|-------------------|
| `GET /_ping` | Returns `OK` |
| `GET /version` | `container system version` |
| `POST /containers/create` | `container run` (deferred start) |
| `POST /containers/{id}/start` | `container start` |
| `POST /containers/{id}/stop` | `container stop` |
| `POST /containers/{id}/kill` | `container kill` |
| `DELETE /containers/{id}` | `container delete` |
| `GET /containers/json` | `container list --format json` |
| `GET /containers/{id}/json` | `container inspect --format json` |
| `GET /containers/{id}/logs` | `container logs` |
| `POST /containers/{id}/exec` | `container exec` |
| `GET /images/json` | `container image list --format json` |
| `POST /images/create` (pull) | `container image pull` |
| `POST /build` | `container build` |
| `POST /networks/create` | `container network create` |
| `DELETE /networks/{id}` | `container network delete` |
| `GET /networks` | `container network list` |

## Scope

gangway implements the **subset** of the Docker Engine API needed by common tools. It does not aim for full Docker API compatibility. Unsupported endpoints return `501 Not Implemented`.

Priority clients:
- OpenClaw sandbox (container create/start/stop/exec/logs/remove)
- Node.js dockerode library
- Testcontainers
- Basic `docker` CLI commands via `DOCKER_HOST`

## How It Works

```
Docker client (OpenClaw, dockerode, etc.)
  |
  | Docker Engine API over unix socket
  |
  v
gangway (Go binary)
  |
  | os/exec: container run/stop/exec/logs/...
  |
  v
Apple container CLI (/opt/homebrew/bin/container)
  |
  | Mach IPC
  |
  v
containermanagerd (macOS system daemon)
  |
  v
Virtualization.framework (lightweight Linux VMs)
```

## Configuration

| Flag | Env | Default | Description |
|------|-----|---------|-------------|
| `--socket` | `GANGWAY_SOCKET` | `/tmp/gangway.sock` | Unix socket path |
| `--container-bin` | `GANGWAY_CONTAINER_BIN` | `container` | Path to Apple container CLI |
| `--log-level` | `GANGWAY_LOG_LEVEL` | `info` | Log verbosity |

## Limitations

- Linux containers only (Apple containers run Linux VMs)
- No Windows container support
- No Docker Compose compatibility (use skiff for multi-container orchestration)
- No swarm/cluster features
- Volume semantics may differ slightly from Docker
- Network model follows Apple container networking, not Docker bridge

## Part of the Agentic Coding Stack

| Tool | Role |
|------|------|
| **skiff** | Container orchestration (lifecycle, health, DNS) |
| **gangway** | Docker API compatibility shim for Apple containers |
| **hoist** | Git worktree management |
| **bosun** | Agent entrypoint/coordinator |
| **beads** | Task tracking and priority |
| **agent-mail** | Inter-agent messaging and file leases |
