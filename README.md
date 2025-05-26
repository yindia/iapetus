# iapetus 🚀

[![Go Reference](https://pkg.go.dev/badge/github.com/yindia/iapetus.svg)](https://pkg.go.dev/github.com/yindia/iapetus)
[![Go Report Card](https://goreportcard.com/badge/github.com/yindia/iapetus)](https://goreportcard.com/report/github.com/yindia/iapetus)
[![codecov](https://codecov.io/gh/yindia/iapetus/graph/badge.svg?token=6S99FUSPOC)](https://codecov.io/gh/yindia/iapetus)


> ⚠️ **This project is under heavy development. APIs may change frequently. Use with caution in production environments.**

**A robust, extensible Go package for orchestrating and validating command-line workflows with parallel DAG execution, built-in assertions, and advanced observability.**

---

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Defining Tasks](#defining-tasks)
- [Workflow Orchestration](#workflow-orchestration)
- [Assertions](#assertions)
- [Advanced Usage](#advanced-usage)
- [Extensibility](#extensibility)
- [YAML/Config Integration](#yamlconfig-integration)
- [Testing & Reliability](#testing--reliability)
- [Contributing](#contributing)
- [License](#license)

---

## Features

- ⚡ **Parallel, dependency-aware workflow execution** (DAG-based)
- ✅ **Built-in and custom assertions** for validating outputs
- 🔄 **Retries** for flaky operations
- 🏗️ **Fluent builder pattern** for readable workflow and task construction
- 🔍 **Observability hooks** and pluggable logging (uses [zap](https://github.com/uber-go/zap))
- 🧪 **Battle-tested**: stress, property-based, and concurrency tests
- 🐳 **Container-ready**: Support for specifying container images and environment variables
- 🧩 **Extensible**: Add custom hooks, assertions, and workflow logic

---

## Installation

```sh
go get github.com/yindia/iapetus
```

---

## Quick Start

Here's how to get started with iapetus in just a few lines of Go:

```go
package main

import (
    "github.com/yindia/iapetus"
    "go.uber.org/zap"
    "time"
)

func main() {
    // Define a task that runs a shell command and asserts its output
    task := iapetus.NewTask("hello-world", 2*time.Second, nil).
        AddCommand("echo").
        AddArgs("Hello, world!").
        AssertOutputContains("Hello")

    // Create a workflow and add the task
    workflow := iapetus.NewWorkflow("quickstart", zap.NewExample()).
        AddTask(*task)

    // Run the workflow
    if err := workflow.Run(); err != nil {
        panic(err)
    }
}
```

- This example runs `echo "Hello, world!"` and checks that the output contains "Hello".
- You can add more tasks, assertions, and hooks as needed.

---

## Defining Tasks

You can define tasks using either **struct literals** or the **builder/fluent API**. Both are fully supported and interoperable.

### Struct Literal Style

```go
task := &iapetus.Task{
    Name:    "verify-service",
    Command: "curl",
    Args:    []string{"-f", "http://localhost:8080"},
    Asserts: []func(*iapetus.Task) error{
        iapetus.AssertExitCode(0),
        iapetus.AssertOutputContains("Success"),
    },
    Image:   "alpine:3.18", // Optional: for containerized execution
    Env:     []string{"FOO=bar"},
    EnvMap:  map[string]string{"FOO": "bar"},
}
```

### Builder/Fluent API Style

```go
task := iapetus.NewTask("verify-service", 5*time.Second, nil).
    AddCommand("curl").
    AddArgs("-f", "http://localhost:8080").
    AssertExitCode(0).
    AssertOutputContains("Success").
    AddImage("alpine:3.18").
    AddEnv("FOO=bar").
    AddEnvMap(map[string]string{"FOO": "bar"})
```

- **Preferred:** Use the new assertion functions that accept expected values directly.
- Use the fluent `Expect()` DSL for readable, chainable assertions.

---

## Workflow Orchestration

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

## Assertions

### Built-in Assertions

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

## Advanced Usage

- **Retries**: Set `Retries` on a task to automatically retry on assertion failure.
- **Timeouts**: Set `Timeout` for each task to prevent hangs.
- **Container Image**: Use the `Image` field to specify a container image (for future containerized runners).
- **Environment Variables**: Use `Env` (list) or `EnvMap` (map) for environment configuration.
- **PreRun/PostRun Hooks**: Use `PreRun` and `PostRun` for setup/teardown logic at both workflow and task levels.

---

## Extensibility

- **Hooks**: Register multiple hooks for task start, success, failure, and completion:

```go
workflow.AddOnTaskStartHook(func(task *iapetus.Task) { /* ... */ })
workflow.AddOnTaskSuccessHook(func(task *iapetus.Task) { /* ... */ })
workflow.AddOnTaskFailureHook(func(task *iapetus.Task, err error) { /* ... */ })
workflow.AddOnTaskCompleteHook(func(task *iapetus.Task) { /* ... */ })
```