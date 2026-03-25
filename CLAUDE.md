# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is the **cluster-logging must-gather** tool for OpenShift - a diagnostic collection utility built on top of OpenShift's must-gather framework. It gathers comprehensive diagnostic information about the OpenShift Cluster Logging subsystem for troubleshooting and support.

## Repository Structure

The repository uses git worktrees (`.bare` directory structure). The main codebase is in `must-gather/collection-scripts/`:

- `gather` - Main entry point script that orchestrates all collection
- `common` - Shared utility functions (logging, environment extraction)
- `gather_cluster_logging_operator_resources` - Collects CLO-specific resources
- `gather_collection_resources` - Collects log collector (Vector) resources
- `gather_logstore_resources` - Collects Elasticsearch or Lokistack resources
- `gather_monitoring` - Collects Prometheus alerts and monitoring data
- `monitoring_common.sh` - Shared monitoring utilities

## How must-gather Works

When run via `oc adm must-gather --image=quay.io/openshift-logging/cluster-logging-operator:latest -- /usr/bin/gather`, the tool:

1. Creates a pod in the cluster with the must-gather image
2. Executes the `/usr/bin/gather` script inside the pod
3. Collects diagnostic data using `oc` commands and direct pod exec
4. Organizes output into a structured directory hierarchy
5. Downloads the collected data to the local machine

## Collection Architecture

### Main Collection Flow

The `gather` script coordinates parallel collection of:

1. **Cluster-scoped resources**: nodes, clusterroles, persistentvolumes, CRDs
2. **Namespace-scoped resources**: All resources in openshift-logging and multi-forwarder namespaces
3. **Component-specific data**: CLO, collectors, log stores, monitoring

### Multi-forwarder Support

The tool automatically discovers namespaces with ClusterLogForwarder CRs and collects their resources:
- Searches for all `clusterlogforwarder` CRDs
- Iterates through namespaces containing these CRs
- Collects collector pods, daemonsets, and configuration for each namespace

### Output Structure

```
must-gather/
├── cluster-logging/
│   ├── clo/                          # CLO operator info and version
│   └── namespaces/[namespace]/       # Per-namespace collector data
├── cluster-scoped-resources/          # Nodes, PVs, CRDs, etc.
├── namespaces/[namespace]/            # Full namespace dumps (via oc adm inspect)
├── monitoring/prometheus/             # Alert rules and Prometheus data
└── gather-debug.log                   # Collection log file
```

## Key Technical Details

### Script Conventions

- All scripts use `set -euo pipefail` for safety
- Background jobs tracked via `pids=()` array, synchronized with `wait`
- Logging uses `log()` function from `common` with timestamps
- Cache directory (`KUBECACHEDIR`) used to avoid repeated API calls

### Resource Collection Methods

- **oc adm inspect**: Primary method for collecting Kubernetes resources (YAML dumps)
- **oc describe**: Used for collector pods and daemonsets
- **oc exec**: Used to extract configs (e.g., vector.toml) and Elasticsearch cluster state
- **curl**: Used inside pods to query Elasticsearch and Prometheus APIs

### Component-Specific Logic

**Elasticsearch** (when present):
- Queries cluster health, indices, nodes via REST API
- Collects additional diagnostics if cluster health is not green
- Lists persistence storage and size information

**Lokistack** (when present):
- Collects Lokistack CRs via `oc adm inspect`

**Collectors**:
- Describes daemonsets and pods
- Extracts Vector configuration from configmap data

## Development Notes

- This repository contains only collection scripts, not the full operator codebase
- Scripts must be compatible with the must-gather pod environment
- All file paths are relative to `BASE_COLLECTION_PATH` (default: `/must-gather`)
- The tool is designed for x86_64 architecture only

## Testing Changes

When modifying collection scripts:
1. Test in a live OpenShift cluster with logging installed
2. Run: `oc adm must-gather --image=<your-test-image> -- /usr/bin/gather`
3. Verify output directory structure matches expected format
4. Check `gather-debug.log` for errors
5. Ensure tools like `omc` can still parse the output

## Common Operations

### Checking what resources are collected

Look at the namespace_resources and cluster_resources arrays in the `gather` script.

### Adding new resource collection

1. Add to appropriate gather_* script or create new specialized script
2. Update `gather` to call your script
3. Ensure output goes to appropriate subdirectory under `BASE_COLLECTION_PATH`
4. Add logging with `log` function
5. Handle errors gracefully (resources may not exist in all clusters)
