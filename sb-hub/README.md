# sb-hub

A command-line tool for spinning up, managing, and tearing down isolated development sandboxes backed by Docker containers.

Built because juggling `docker run` commands, port conflicts, and leftover volumes across multiple projects gets old fast. sb-hub wraps the Docker API with opinionated defaults — auto-assigned ports, size presets, TTL-based expiry, persistent storage, and snapshot/restore — so you can focus on the work instead of the plumbing.

---

## Architecture

```
┌──────────────────────────────────────────────────────────┐
│                        CLI Layer                         │
│                     (cobra commands)                     │
│                                                          │
│   create  list  remove  console  logs  save  renew ...   │
└────────────────────────┬─────────────────────────────────┘
                         │
                         ▼
┌──────────────────────────────────────────────────────────┐
│                    Engine (pkg/)                          │
│                                                          │
│  Dockerengine ──► DockerClient interface                  │
│                                                          │
│  • Container lifecycle (create, start, stop, remove)     │
│  • Image management (pull, build)                        │
│  • Network management (sb-hub-net bridge)                │
│  • TTL tracking via container labels                     │
│  • Port discovery                                        │
└────────────────────────┬─────────────────────────────────┘
                         │
                         ▼
┌──────────────────────────────────────────────────────────┐
│                     Docker Daemon                         │
│                                                          │
│  ┌───────────┐  ┌───────────┐  ┌───────────┐            │
│  │ sandbox-a │  │ sandbox-b │  │ sandbox-c │  ...        │
│  │  :8001    │  │  :8002    │  │  :8003    │             │
│  └─────┬─────┘  └─────┬─────┘  └─────┬─────┘            │
│        └───────────────┼───────────────┘                 │
│                   sb-hub-net                              │
│                  (bridge network)                         │
└──────────────────────────────────────────────────────────┘
                         │
                         ▼
┌──────────────────────────────────────────────────────────┐
│                   Host Filesystem                         │
│                                                          │
│  storage-root/                                           │
│  ├── sandbox-a/              ← live data (bind mount)    │
│  ├── sandbox-b/                                          │
│  ├── sandbox-a_snap_v1/      ← manual snapshot           │
│  └── sandbox-b_janitor_.../  ← auto-archived by janitor  │
└──────────────────────────────────────────────────────────┘
```

### How the pieces fit together

The **CLI layer** parses flags and arguments, then hands off to the **Engine**. The engine talks to the Docker daemon through a `DockerClient` interface (a thin wrapper around the Docker SDK). Every sandbox is a regular Docker container with a few custom labels stamped on it:

| Label | Purpose |
|---|---|
| `com.sbhub.managed` | Marks the container as sb-hub–managed |
| `com.sbhub.expires` | RFC 3339 timestamp for TTL expiry |
| `com.sbhub.size` | Size preset used to create it |
| `com.sbhub.hostport` | The auto-assigned host port |

The **janitor** process reads these labels to decide what's expired, then archives and removes stale containers automatically.

---

## What it does

### Sandbox lifecycle

```
create ──► running ──► save (snapshot) ──► remove
              │                               │
              ├── renew (extend TTL)          │
              ├── console (shell in)          │
              ├── logs (stream output)        │
              └── janitor (auto-expire) ──────┘
```

**Creating** a sandbox picks a size preset, pulls the base image, finds an available port, creates a data directory on the host, and starts the container with everything wired up. You get a running environment with persistent storage and a URL to hit.

**Saving** copies the live data directory to a tagged snapshot. **Restoring** copies it back when creating a new sandbox. This lets you checkpoint your work and roll back if needed.

The **janitor** runs as a background loop, checking every 30 seconds for containers whose TTL has passed. When it finds one, it stops the container and moves the data to an archive directory rather than deleting it outright.

### Size presets

| Preset | CPU | Memory | Disk | Default TTL |
|---|---|---|---|---|
| small | 0.5 cores | 512 MB | 10 GB | 6 hours |
| medium | 2 cores | 4 GB | 20 GB | 4 hours |
| large | 4 cores | 8 GB | 40 GB | 2 hours |
| xlarge | 8 cores | 16 GB | 80 GB | 1 hour |

Larger sandboxes get shorter TTLs by default — the idea is that heavier environments shouldn't linger if you forget about them. You can override the TTL at creation time.

### Networking

All sandboxes land on a shared `sb-hub-net` bridge network, which means they can talk to each other by container name. Each sandbox also gets a host port mapped automatically from the 8000–9000 range, so you can access services from your browser without hunting for ports.

### Project imports

Point `sb import` at a project directory and it does the right thing:

- **Has a Dockerfile?** Builds a custom image and launches it as a sandbox.
- **Has a docker-compose.yml?** Parses the services and spins up a sandbox for each one, all on the shared network so they can discover each other.

### Storage operations

| Command | What it does |
|---|---|
| `save` | Snapshot current data to a tagged copy |
| `attach` | Hot-swap a sandbox to a different data folder |
| `detach` | Remove all mounts, make a sandbox stateless |

---

## Project structure

```
sb-hub/
├── main.go              # Entry point — just calls cmd.Execute()
├── cmd/
│   ├── root.go          # Base cobra command
│   ├── create.go        # Create sandbox with auto-port and size presets
│   ├── list.go          # List active + archived sandboxes
│   ├── remove.go        # Tear down sandbox and wipe data
│   ├── console.go       # Interactive shell into a sandbox
│   ├── logs.go          # Stream container logs
│   ├── save.go          # Snapshot sandbox data
│   ├── renew.go         # Extend TTL
│   ├── attach.go        # Switch data folder
│   ├── detach.go        # Remove data mounts
│   ├── import.go        # Import Dockerfile/Compose projects
│   └── janitor.go       # Background TTL enforcer
├── pkg/
│   ├── engine.go        # Docker API wrapper + DockerClient interface
│   └── types.go         # Size presets and sandbox specs
└── tests/
    ├── engine_test.go   # Engine tests with mock Docker client
    ├── types_test.go    # Sandbox spec validation
    ├── create_test.go   # Port selection logic
    └── import_test.go   # Compose YAML parsing
```

---

## Commands at a glance

| Command | Description |
|---|---|
| `sb create [name]` | Spin up a new sandbox |
| `sb list` | Show all sandboxes and archived data |
| `sb remove [name]` | Tear down a sandbox |
| `sb console [name]` | Shell into a running sandbox |
| `sb logs [name]` | View container output |
| `sb save [name] [tag]` | Snapshot sandbox data |
| `sb renew [name] [duration]` | Extend the TTL |
| `sb attach [name] [folder]` | Switch to a different data folder |
| `sb detach [name]` | Remove storage mounts |
| `sb import [path]` | Import a Dockerfile or Compose project |
| `sb janitor` | Start the background TTL enforcer |

---

## Testing

The test suite uses a `MockDockerClient` that implements the `DockerClient` interface, so tests run instantly without needing a real Docker daemon. Tests cover the engine layer (container lifecycle, networking, port discovery, TTL expiry logic), sandbox spec validation, port selection edge cases, and Compose YAML parsing.

```
$ go test ./tests/ -v
--- 43 tests, all passing ---
```

---

## License

MIT
