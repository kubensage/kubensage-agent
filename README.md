# kubensage-agent

`kubensage-agent` is a lightweight telemetry agent designed to run on Kubernetes nodes. It collects detailed host- and
container-level metrics using the Kubernetes CRI (Container Runtime Interface) by establishing a direct gRPC connection
to the container runtime socket. The collected data is sent via GRPC to an external relay for enrichment and export to
systems like Prometheus.

---

## ✨ Features

* Collects **node metrics**: CPU, memory, PSI, network interfaces, uptime, kernel info, etc.
* Collects **pod & container metrics**: per-container CPU, memory, filesystem, and swap usage.
* Connects directly to the **CRI runtime socket** using **gRPC** to gather runtime data from `containerd`, `CRI-O`, or
  `dockershim`.
* Clean, modular Go code with extensibility in mind.

---

## 📊 Metrics Overview

The agent collects a wide variety of metrics structured into the following main objects:

### `Metrics`

Top-level struct containing everything gathered during one polling iteration.

```go
package metrics

type Metrics struct {
	NodeMetrics *NodeMetrics
	PodMetrics  []*PodMetrics
}

```

### `NodeMetrics`

Host-level system metrics: CPU info, memory usage, PSI, network interfaces, OS/kernel metadata.

### `PodMetrics` & `ContainerMetrics`

Each pod includes container-level statistics, such as:

* CPU usage (nano cores / core-seconds)
* Memory usage (RSS, WorkingSet, etc.)
* Filesystem usage (used bytes, inodes)
* Swap usage (optional)

All metrics are safely extracted, even in partial or incomplete container states.

---

## 🔄 Agent Loop

The `main.go` runs a periodic collection loop:

1. Detects the CRI socket (containerd, CRI-O, dockershim)
2. Opens a **gRPC connection** to the container runtime via the detected socket (e.g.,
   `/run/containerd/containerd.sock`)
3. Every `n seconds`, calls `discovery.GetAllMetrics()` to collect all available data
4. Converts the metrics to Proto
5. Sends them to the relay, via GRPC

Handles SIGINT/SIGTERM gracefully.

---

## 🛠️ Building the Agent

To build the `kubensage-agent` binary for multiple platforms, a `Makefile` is provided.
The output binaries are placed in `.go-builds/`, and versioning is controlled via the `VERSION` variable.

### 🔧 Available Make Targets

* `make build-all`: Builds binaries for all supported OS/ARCH combinations:

    * Linux (amd64, arm64)
    * macOS (amd64, arm64)
    * Windows (amd64)

* `make build-proto`: Compiles -proto files

* `make build-linux-amd64`: Builds for Linux x86\_64

* `make build-linux-arm64`: Builds for Linux ARM64

* `make build-darwin-amd64`: Builds for macOS Intel

* `make build-darwin-arm64`: Builds for macOS Apple Silicon

* `make build-windows-amd64`: Builds for Windows x86\_64

* `make clean`: Removes the `.go-builds` output directory

### 🏷️ Release

To tag a release in Git you just need to open a PR to main branch, the CI will create automatically a new TAG and a new
Release.

### 🧪 Example

```bash
VERSION=1.0.0 make build-linux-amd64
```

This builds `kubensage-agent-1.0.0-linux-amd64` in the `.go-builds/` directory.

Ensure Go is installed and in your `PATH`, and that you are in the root of the repository when running `make` commands.

Logs are written to `kubensage-agent.log` in append mode. Ensure the agent has read access to CRI socket (usually root).

---

## 📡 Roadmap

* [x] Collect node + container metrics
* [x] Structured logging
* [x] PSI support
* [x] Agent push to relay
* [ ] Systemd service unit for agent deployment

---

## 📄 License (for now)

MIT © 2025 kubensage authors
