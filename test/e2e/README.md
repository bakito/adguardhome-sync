# End-to-End (E2E) Testing

This directory contains the Go end-to-end test suite for `adguardhome-sync`.

## Overview & Architecture

All E2E testing is executed against a Kubernetes cluster (e.g. [Kind](https://kind.sigs.k8s.io/)). The Helm chart in `testdata/e2e` deploys:
- **Origin AdGuardHome** instance (`adguardhome-origin`) exposing NodePort `30000` (mapped to host port `3000`).
- **3 Replica AdGuardHome** instances of varying versions:
  - Replica 1 (`v0.107.68`): NodePort `30001` (mapped to host port `9091`).
  - Replica 2 (`v0.107.79`): NodePort `30002` (mapped to host port `9092`).
  - Replica 3 (`latest`): NodePort `30003` (mapped to host port `9093`).
- **adguardhome-sync**: can run either inside the cluster (port `9090`) or externally on the host (e.g. in an IDE debugger on `localhost:9090`).

When creating the Kind cluster via `make e2e-setup-k8s`, Kind uses `testdata/e2e/kind.yaml` to map host ports directly to Kubernetes NodePort services for origin and replica instances. This eliminates the need for background port-forwarding processes.

---

## Accessing AdGuardHome Instances & Web UI

When the Kind cluster is running (`make e2e-setup-k8s`), you can access the Web UI and REST API of each AdGuardHome instance directly from your browser or HTTP client using host ports:

| Instance | Version | Host URL (Browser / API) | In-Cluster Service DNS | NodePort | Credentials (User / Pass) |
|---|---|---|---|---|---|
| **Origin** | `latest` | [http://localhost:3000](http://localhost:3000) | `service-origin.agh-e2e.svc.cluster.local:3000` | `30000` | `username` / `password` |
| **Replica 1** | `v0.107.68` | [http://localhost:9091](http://localhost:9091) | `service-replica-v0-107-68.agh-e2e.svc.cluster.local:3000` | `30001` | `username` / `password` |
| **Replica 2** | `v0.107.79` | [http://localhost:9092](http://localhost:9092) | `service-replica-v0-107-79.agh-e2e.svc.cluster.local:3000` | `30002` | `username` / `password` |
| **Replica 3** | `latest` | [http://localhost:9093](http://localhost:9093) | `service-replica-latest.agh-e2e.svc.cluster.local:3000` | `30003` | `username` / `password` |

### adguardhome-sync Endpoints (when running)
- **API Status**: [http://localhost:9090/api/v1/status](http://localhost:9090/api/v1/status)
- **Prometheus Metrics**: [http://localhost:9090/metrics](http://localhost:9090/metrics)

---

## Updating Replica Versions

Replica versions tested in E2E suites are configured in a single place: [`testdata/e2e/values.yaml`](../../testdata/e2e/values.yaml).

You can update them automatically to match the target AdGuardHome version (`ADGUARD_HOME_VERSION` in Makefile):
```bash
make update-e2e-versions
```
Or with a specific version:
```bash
make update-e2e-versions ADGUARD_HOME_VERSION=v0.107.80
```
Running `make generate` also triggers `update-e2e-versions` automatically to keep `values.yaml` in sync.

---

## 1. Running Full In-Cluster E2E Tests (Kind / CI)

### Step 1: Set up Local Kubernetes (Kind) Cluster & Deploy Chart
Creates a local kind cluster, builds the container image, loads it into the cluster, and installs the helm chart:
```bash
make e2e-setup-k8s
```
*(Optionally specify `MODE=file` or custom cluster name `KIND_CLUSTER_NAME=...`)*

### Step 2: Run the E2E Test Suite
```bash
make test-e2e
```

### Step 3: Cleanup Local Kubernetes Cluster
```bash
make e2e-stop-k8s
```

---

## 2. Debugging `adguardhome-sync` Outside the Cluster (in IDE)

You can run `adguardhome-sync` locally in your IDE with active breakpoints while synchronizing against the AdGuardHome instances running in the Kubernetes cluster.

### Step 1: Set up the Cluster (Without In-Cluster Sync)
Deploy the Kind cluster and install only the AdGuardHome origin and replica instances:
```bash
make e2e-setup-k8s SYNC_ENABLED=false
```

### Step 2: Run / Debug in IDE
Use the provided sample configuration `testdata/e2e/local-config.yaml`.

#### In GoLand / IntelliJ:
1. Open **Run/Debug Configurations**.
2. Add a new **Go Build** configuration:
   - **Package path**: `github.com/bakito/adguardhome-sync` (or directory `.`)
   - **Program arguments**: `run --config testdata/e2e/local-config.yaml`
   - **Working directory**: Project root (`adguardhome-sync`)
3. Place breakpoints in any file (e.g. `internal/sync/synchronizer.go`, `cmd/run.go`).
4. Click **Debug**.

#### In VS Code:
Create or use `.vscode/launch.json`:
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug adguardhome-sync",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/main.go",
      "args": ["run", "--config", "testdata/e2e/local-config.yaml"]
    }
  ]
}
```

### Step 3: Run the E2E Test Suite Against the IDE Process
In your terminal, execute the test suite in external sync mode:
```bash
make test-e2e-external
```
Or directly using `go test`:
```bash
E2E_EXTERNAL_SYNC=true go test -v -tags=e2e ./test/e2e/...
```

The test suite will validate sync status, Prometheus metrics, healthcheck commands, replica feature overrides, and replica configuration against the cluster and your debugged process.

### Step 4: Cleanup
```bash
make e2e-stop-k8s
```

---

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `E2E_EXTERNAL_SYNC` | `false` | Set to `true` when testing against an external/IDE sync process |
| `E2E_PROTOCOL` | `https` (env) / `http` (file/external) | Protocol (`http` or `https`) |
| `E2E_MODE` | `env` | Sync mode (`env` or `file`) |
| `E2E_SYNC_URL` | `http://localhost:9090` | Sync API base URL |
| `E2E_ORIGIN_URL` | `http://localhost:3000` | Origin AdGuardHome URL |
| `E2E_REPLICA_URLS` | `http://localhost:9091,...` | Comma-separated replica URLs |
| `E2E_NAMESPACE` | `agh-e2e` | Kubernetes namespace |
| `E2E_TIMEOUT` | `60s` | Sync completion timeout |
| `E2E_LOCAL_CONFIG_FILE` | `testdata/e2e/local-config.yaml` | Config file path for local healthcheck tests |
