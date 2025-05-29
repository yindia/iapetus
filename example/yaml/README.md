# YAML Workflow Examples for iapetus

This directory demonstrates how to define and run iapetus workflows using YAML configuration files.

## How it works
- Workflows are defined in YAML (see `workflow_docker.yaml`).
- Assertions are specified in the `raw_asserts` field for each task.
- The Go loader (`iapetus.LoadWorkflowFromYAML`) parses the YAML, converts assertions, and returns a ready-to-run workflow.

## Running the Docker YAML Example

1. Make sure you have Docker installed and available in your PATH.
2. Run the example:

```sh
cd example/yaml
# Make sure dependencies are installed (from repo root):
# go mod tidy
# Run the YAML workflow example:
go run main_docker.go
```

You should see output for each task, and the workflow will complete successfully if all assertions pass.

## Pattern
- Use `raw_asserts` in YAML for all supported assertion types.
- No post-processing is needed after loading: the loader handles conversion.
- See the main README for more details and advanced usage. 