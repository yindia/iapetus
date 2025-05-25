# iapetus 🚀

[![Go Reference](https://pkg.go.dev/badge/github.com/yindia/iapetus.svg)](https://pkg.go.dev/github.com/yindia/iapetus)
[![Go Report Card](https://goreportcard.com/badge/github.com/yindia/iapetus)](https://goreportcard.com/report/github.com/yindia/iapetus)

A robust, extensible Go package for orchestrating and validating command-line workflows with parallel DAG execution, built-in assertions, and advanced observability.

---

## Overview 📋

- ⚡ **Parallel, dependency-aware workflow execution** (DAG-based)
- ✅ **Built-in and custom assertions** for validating outputs
- 🔄 **Retries** for flaky operations
- 🏗️ **Fluent builder pattern** for readable workflow and task construction
- 🔍 **Observability hooks** and pluggable logging (uses [zap](https://github.com/uber-go/zap))
- 🧪 **Battle-tested**: stress, property-based, and concurrency tests

---

## Installation 💻

```sh
go get github.com/yindia/iapetus
```

---

## Quick Start

```go
import (
    "time"
    "log"
    "github.com/yindia/iapetus"
    "go.uber.org/zap"
)

// Create a simple task
step := iapetus.NewTask("verify-service", 5*time.Second, nil).
    AddCommand("curl").
    AddArgs("-f", "http://localhost:8080").
    AddExpected(iapetus.Output{ExitCode: 0}).
    AddAssertion(iapetus.AssertByExitCode)

// Run the task
if err := step.Run(); err != nil {
    log.Fatalf("Task failed: %v", err)
}
```

---

## Defining Tasks

A `Task` represents a command to run, its arguments, environment, expected output, and assertions.

```go
task := iapetus.NewTask("kubectl-get-pods", 10*time.Second, nil).
    AddCommand("kubectl").
    AddArgs("get", "pods", "-n", "default").
    AddExpected(iapetus.Output{ExitCode: 0}).
    AddAssertion(iapetus.AssertByExitCode)
```

- **Assertions**: Add built-in or custom assertions with `AddAssertion`.
- **Retries**: Use `SetRetries(n)` to retry on failure.
- **Environment**: Use `AddEnv` to set environment variables.

---

## Orchestrating Workflows (DAG)

A `Workflow` is a collection of tasks with dependencies, executed in parallel where possible (DAG scheduling).

```go
workflow := iapetus.NewWorkflow("cluster-setup", zap.NewNop()).
    AddTask(*step1).
    AddTask(*step2).
    AddPreRun(func(w *iapetus.Workflow) error {
        // Setup logic before tasks run
        return nil
    }).
    AddPostRun(func(w *iapetus.Workflow) error {
        // Cleanup logic after all tasks
        return nil
    })

if err := workflow.Run(); err != nil {
    log.Fatalf("Workflow failed: %v", err)
}
```

- **Parallelism**: Tasks run in parallel, respecting dependencies (`Depends` field).
- **Hooks**: Register hooks for task start, success, failure, and completion.
- **Observability**: Pluggable logger (defaults to zap), all events are logged.

---

## Assertions

Built-in assertion functions:

- `AssertByExitCode`: Validates the exit code
- `AssertByOutputString`: Exact string match
- `AssertByOutputJson`: JSON comparison (with node skipping)
- `AssertByContains`: Substring presence
- `AssertByError`: Error message validation (substring or regexp)
- `AssertByRegexp`: Output matches regular expression

**Example:**

```go
task.AddAssertion(iapetus.AssertByExitCode).
     AddAssertion(iapetus.AssertByContains("Ready"))
```

---

## Custom Assertions

Add your own validation logic:

```go
task.AddAssertion(func(t *iapetus.Task) error {
    if !strings.Contains(t.Actual.Output, "success") {
        return fmt.Errorf("expected 'success' in output")
    }
    return nil
})
```

---

## Advanced Features

- **Parallel DAG scheduling**: Tasks run as soon as dependencies are met
- **Observability hooks**: Register multiple listeners for task lifecycle events
- **Stress-tested**: Handles thousands of tasks, deep dependency chains
- **Property-based testing**: Ensures correctness for random DAGs and workflows
- **Pluggable logging**: Uses [zap](https://github.com/uber-go/zap) by default, can be replaced
- **Extensible**: Add custom hooks, assertions, and workflow logic

---

## Extensibility

- **Hooks**: Register multiple hooks for task start, success, failure, and completion:

```go
workflow.AddOnTaskStartHook(func(task *iapetus.Task) { /* ... */ })
workflow.AddOnTaskSuccessHook(func(task *iapetus.Task) { /* ... */ })
workflow.AddOnTaskFailureHook(func(task *iapetus.Task, err error) { /* ... */ })
workflow.AddOnTaskCompleteHook(func(task *iapetus.Task) { /* ... */ })
```

- **Custom Pre/Post Run**: Use `AddPreRun` and `AddPostRun` for workflow-level setup/teardown.

---

## Contributing & Testing

- **Contributions welcome!** Please open issues or PRs.
- **Testing**: Run all tests (including stress and property-based):

```sh
go test -v ./...
```

---

## License 📄

MIT License. See [LICENSE](LICENSE) for details.

