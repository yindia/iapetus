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

## Defining Tasks: Two Approaches

You can define tasks using either **struct literals** or the **builder/fluent API**. Both are fully supported and interoperable.

### 1. Struct Literal Style

```go
import "github.com/yindia/iapetus"

task := &iapetus.Task{
    Name:    "verify-service",
    Command: "curl",
    Args:    []string{"-f", "http://localhost:8080"},
    Asserts: []func(*iapetus.Task) error{
        iapetus.AssertExitCode(0),
        iapetus.AssertOutputContains("Success"),
    },
}
```

- **Preferred:** Use the new assertion functions that accept expected values directly.
- No need to set `Expected` fields for assertions.

### 2. Builder/Fluent API Style

```go
import "github.com/yindia/iapetus"

task := iapetus.NewTask("verify-service", 5*time.Second, nil).
    AddCommand("curl").
    AddArgs("-f", "http://localhost:8080").
    AssertExitCode(0).
    AssertOutputContains("Success").
    AssertOutputEquals("expected output").
    AssertOutputJsonEquals(`{"foo":"bar"}`).
    AssertOutputMatchesRegexp("pattern").
    Expect().OutputContains("Success").ExitCode(0).Done()
```

- Pass expected values directly to assertion methods—no need to set `Expected` fields.
- Use the fluent `Expect()` DSL for readable, chainable assertions.

---

## Assertion Usage: Summary Table

| Approach         | Example Usage                                                                 | When to Use                |
|------------------|-------------------------------------------------------------------------------|----------------------------|
| Struct Literal   | `Asserts: []func(*Task) error{AssertExitCode(0), AssertOutputContains("foo")}`| Static config, YAML, tests |
| Builder/Fluent   | `.AssertExitCode(0).AssertOutputContains("foo")`                             | Programmatic, chaining     |
| Assertion DSL    | `.Expect().ExitCode(0).OutputContains("foo").Done()`                        | Complex, readable chains   |

- **All approaches are interoperable**: you can mix and match as needed.
- **Preferred:** Use the new assertion functions with expected values as arguments.

---

## Orchestrating Workflows (DAG)

A `Workflow` is a collection of tasks with dependencies, executed in parallel where possible (DAG scheduling).

```go
workflow := iapetus.NewWorkflow("cluster-setup", zap.NewNop()).
    AddTask(*task1).
    AddTask(*task2).
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

## Built-in Assertions

- `AssertExitCode(expected int)`: Validates the exit code
- `AssertOutputEquals(expected string)`: Exact string match
- `AssertOutputJsonEquals(expected string, skipJsonNodes ...string)`: JSON comparison
- `AssertOutputContains(substr string)`: Substring presence
- `AssertOutputMatchesRegexp(pattern string)`: Output matches regular expression

### Custom Assertions

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

