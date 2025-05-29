# Docker Backend Example for iapetus

This directory demonstrates how to use the iapetus workflow engine with the Docker backend.

## What it shows
- How to define and run workflows that execute commands inside Docker containers.
- How to set environment variables, images, and assertions for each task.
- How to use Go code to define workflows (see `main.go`).

## Running the Go-based Docker Example

1. Make sure you have Docker installed and available in your PATH.
2. Run the example:

```sh
cd example/docker
# Make sure dependencies are installed (from repo root):
# go mod tidy
# Run the Go workflow example:
go run main.go
```

You should see output for each task, and the workflow will complete successfully if all assertions pass.

## YAML-based Workflows

For YAML-based workflow examples (including Docker), see the [example/yaml](../yaml) directory. 