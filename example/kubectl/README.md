# Kubectl Backend Example for iapetus

This directory demonstrates how to use the iapetus workflow engine to orchestrate Kubernetes resources using the kubectl CLI.

## What it shows
- How to define and run workflows that manage Kubernetes clusters, namespaces, and deployments.
- How to use Go code to define complex workflows with dependencies, environment variables, and assertions (see `main.go`).
- How to perform integration/E2E tests for Kubernetes using iapetus.

## Running the Kubectl Example

1. Make sure you have `kubectl` installed and configured to access a Kubernetes cluster (e.g., kind, minikube, or a real cluster).
2. Run the example:

```sh
cd example/kubectl
# Make sure dependencies are installed (from repo root):
# go mod tidy
# Run the Go workflow example:
go run main.go
```

You should see output for each step, and the workflow will complete successfully if all assertions pass.

## YAML-based Workflows

For YAML-based workflow examples, see the [example/yaml](../yaml) directory. 