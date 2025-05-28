# iapetus 🚀

[![Go Reference](https://pkg.go.dev/badge/github.com/yindia/iapetus.svg)](https://pkg.go.dev/github.com/yindia/iapetus)
[![Go Report Card](https://goreportcard.com/badge/github.com/yindia/iapetus)](https://goreportcard.com/report/github.com/yindia/iapetus)
[![codecov](https://codecov.io/gh/yindia/iapetus/graph/badge.svg?token=6S99FUSPOC)](https://codecov.io/gh/yindia/iapetus)

---

> ⭐ **If you like this project, please [star us on GitHub](https://github.com/yindia/iapetus/stargazers)!**

---

# ⚡ iapetus: The Ultimate Go Workflow Orchestrator

**A robust, extensible Go package for orchestrating and validating command-line workflows with parallel DAG execution, built-in assertions, and advanced observability.**

---

## Table of Contents
- [Why iapetus?](#why-iapetus)
- [Features](#features)
- [Getting Started](#getting-started)
- [Task: Definition & Usage](#task-definition--usage)
- [Workflow: Orchestration & Hooks](#workflow-orchestration--hooks)
- [Assertions: Built-in & Custom](#assertions-built-in--custom)
- [Advanced Features](#advanced-features)
- [Observability & Extensibility](#observability--extensibility)
- [Contributing](#contributing)
- [Community & Support](#community--support)
- [License](#license)

---

## Why iapetus? 🤔

- **Lightning-fast parallel execution** with dependency-aware DAG scheduling
- **Battle-tested reliability**: stress, property-based, and concurrency tests
- **Pluggable observability**: hooks, metrics, and [zap](https://github.com/uber-go/zap) logging
- **Extensible**: Add your own assertions, hooks, and workflow logic
- **Production-ready**: Used in CI/CD, data pipelines, and cloud automation
- **Developer-friendly**: Fluent builder API, YAML/config support, and rich documentation

---

## ⭐ Features

- ⚡ **Parallel, dependency-aware workflow execution** (DAG-based)
- ✅ **Built-in and custom assertions** for validating outputs
- 🔄 **Retries** for flaky operations
- 🏗️ **Fluent builder pattern** for readable workflow and task construction
- 🔍 **Observability hooks** and pluggable logging
- 🧪 **Battle-tested**: stress, property-based, and concurrency tests
- 🐳 **Container-ready**: Support for specifying container images and environment variables
- 🧩 **Extensible**: Add custom hooks, assertions, and workflow logic

---

## 🚀 Getting Started

### 1. Install iapetus

```sh
go get github.com/yindia/iapetus
```

### 2. Run the Example

We recommend starting with a real example. Clone the repo and run the provided example:

```sh
git clone https://github.com/yindia/iapetus.git
cd iapetus
cd example/docker
# (Optional) Initialize a new Go module if you want to experiment
# go mod init example.com/quickstart
# go mod tidy
# Run the example
GO111MODULE=on go run main.go
```

You should see output from the workflow and task execution. Try editing the command or assertions in `main.go` to see how failures are reported!

Or, to use in your own project, see the code sample below and follow the instructions to set up your own Go module.

---

## 🧩 Task: Definition & Usage

A **Task** represents a single command or operation in your workflow. You can define tasks using struct literals or the fluent builder API.

### Task Fields
- `Name` (string): Unique identifier for the task
- `Command` (string): The command to execute (e.g., `echo`, `curl`)
- `Args` ([]string): Command-line arguments
- `Timeout` (time.Duration): Maximum execution time
- `Retries` (int): Number of retry attempts on failure
- `Depends` ([]string): Names of tasks this task depends on
- `Env` ([]string): Environment variables (KEY=VALUE)
- `EnvMap` (map[string]string): Alternative env representation
- `Image` (string): Container image (for future containerized runners)
- `Asserts` ([]func(*Task) error): List of assertion functions
- `PreRun`/`PostRun` (func): Hooks for setup/teardown

### Task Examples

#### Struct Literal
```go
task := &iapetus.Task{
    Name:    "verify-service",
    Command: "curl",
    Args:    []string{"-f", "http://localhost:8080"},
    Timeout: 5 * time.Second,
    Retries: 2,
    EnvMap:  map[string]string{"FOO": "bar"},
    Depends: []string{"setup-db"},
    Asserts: []func(*iapetus.Task) error{
        iapetus.AssertExitCode(0),
        iapetus.AssertOutputContains("Success"),
    },
    PreRun: func(t *iapetus.Task) error {
        // Custom setup
        return nil
    },
    PostRun: func(t *iapetus.Task) error {
        // Custom teardown
        return nil
    },
}
```

#### Builder/Fluent API
```go
task := iapetus.NewTask("verify-service", 5*time.Second, nil).
    AddCommand("curl").
    AddArgs("-f", "http://localhost:8080").
    AssertExitCode(0).
    AssertOutputContains("Success").
    AddImage("alpine:3.18").
    AddEnvMap(map[string]string{"FOO": "bar"}).
    SetRetries(2).
    AddAssertion(func(t *iapetus.Task) error {
        if !strings.Contains(t.Actual.Output, "OK") {
            return fmt.Errorf("expected OK in output")
        }
        return nil
    })
```

#### With Hooks and Retries
```go
task := iapetus.NewTask("db-migrate", 10*time.Second, nil).
    AddCommand("sh").
    AddArgs("-c", "./migrate.sh").
    SetRetries(3).
    AddEnv("ENV=prod")
    // Add assertions and hooks as needed
```

---

## 🏗️ Workflow: Orchestration & Hooks

A **Workflow** is a collection of tasks with dependencies, executed in parallel where possible (DAG scheduling).

### Workflow Fields
- `Name` (string): Workflow identifier
- `Steps` ([]Task): List of tasks
- `PreRun`/`PostRun` (func): Workflow-level setup/teardown
- `OnTaskStartHooks`, `OnTaskSuccessHooks`, `OnTaskFailureHooks`, `OnTaskCompleteHooks`: Observability hooks
- `Image`, `EnvMap`: Workflow-wide container/image/env config
- `logger`: Pluggable logger (defaults to zap)

### Workflow Example
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

#### Dependency Management
- Use the `Depends` field in each task to specify dependencies by name.
- The scheduler will execute tasks in parallel where possible, respecting dependencies.

#### Observability Hooks
```go
workflow.AddOnTaskStartHook(func(task *iapetus.Task) { /* ... */ })
workflow.AddOnTaskSuccessHook(func(task *iapetus.Task) { /* ... */ })
workflow.AddOnTaskFailureHook(func(task *iapetus.Task, err error) { /* ... */ })
workflow.AddOnTaskCompleteHook(func(task *iapetus.Task) { /* ... */ })
```

---

## 🧪 Assertions: Built-in & Custom

### Built-in Assertions
- `AssertExitCode(expected int)`: Validates the exit code
- `AssertOutputEquals(expected string)`: Exact string match
- `AssertOutputJsonEquals(expected string, skipJsonNodes ...string)`: JSON comparison
- `AssertOutputContains(substr string)`: Substring presence
- `AssertOutputMatchesRegexp(pattern string)`: Output matches regular expression

### Custom Assertions
```go
task.AddAssertion(func(t *iapetus.Task) error {
    if !strings.Contains(t.Actual.Output, "success") {
        return fmt.Errorf("expected 'success' in output")
    }
    return nil
})
```

### Assertion Chaining (Fluent DSL)
```go
task.Expect().ExitCode(0).OutputContains("foo").Done()
```

---

## 🚦 Advanced Features

- **Retries**: Set `Retries` on a task to automatically retry on assertion failure.
- **Timeouts**: Set `Timeout` for each task to prevent hangs.
- **Container Image**: Use the `Image` field to specify a container image (for future containerized runners).
- **Environment Variables**: Use `Env` (list) or `EnvMap` (map) for environment configuration.
- **PreRun/PostRun Hooks**: Use `PreRun` and `PostRun` for setup/teardown logic at both workflow and task levels.
- **Event-driven Scheduler**: iapetus uses an event-driven, concurrency-safe DAG scheduler for robust parallel execution.

---

## 🔍 Observability & Extensibility

- **Logging**: All events are logged with [zap](https://github.com/uber-go/zap) by default. You can provide your own logger.
- **Hooks**: Register multiple hooks for task start, success, failure, and completion (see above).
- **Custom Extensions**: Add your own assertion functions, hooks, or even custom task runners.
- **YAML/Config Integration**: (Planned) Load workflows from YAML or other config formats for CI/CD and automation.

---

## 🤝 Contributing

We welcome contributions! Help us make iapetus the best workflow orchestrator in Go:

### Development Environment Setup

1. **Install [pixi](https://pixi.sh/):**
   - Follow the instructions at https://pixi.sh/ to install pixi on your system.

2. **Enter the development environment:**
   - Run:
     ```sh
     pixi shell
     ```
   - This will drop you into a shell with all dependencies and tools available.

3. **Run development commands with the `pixi` prefix:**
   - For example, to run tests:
     ```sh
     pixi run go test ./...
     ```
   - To run linters:
     ```sh
     pixi run golangci-lint run ./...
     ```

4. **Standard Go commands also work inside `pixi shell`.**

### Code Style & PR Guidelines

- Follow idiomatic Go style (`gofmt`, `golangci-lint`)
- Write clear, descriptive commit messages (conventional commits preferred)
- Add/maintain unit tests for all features and bugfixes
- Document public APIs with GoDoc comments
- Open issues for bugs, features, or improvements
- Be kind and constructive in code reviews and discussions

---

## 🌎 Community & Support

- **Issues:** [File a bug or feature request](https://github.com/yindia/iapetus/issues)
- **Discussions:** [Join the conversation](https://github.com/yindia/iapetus/discussions)
- **Pull Requests:** [Contribute code or docs](https://github.com/yindia/iapetus/pulls)
- **Questions?** Open a discussion or reach out via issues!

---

## 📜 License

[MIT](LICENSE)

---

> ⭐ **If you find iapetus useful, please [star the repo](https://github.com/yindia/iapetus/stargazers) and share it with your friends and colleagues!**
