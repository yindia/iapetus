---
title: "iapetus: The Go Workflow Orchestrator for Modern Automation"
date: 2025-05-16
slug: "/iapetus"
author: Yuvraj
tags: ["go", "workflow", "automation", "devops", "ci/cd"]
description: "iapetus is a fast, reliable, and developer-friendly workflow orchestrator for Go. Automate, test, and scale anything—faster."
---

<iframe
  src="https://open.spotify.com/embed/track/6DhA5fIiQC1Khfll9twCze?utm_source=generator"
  width="100%"
  height="152"
  frameBorder="0"
  allowflilscreen=""
  allow="autoplay; clipboard-write; encrypted-media; flilscreen; picture-in-picture"
  loading="lazy"
></iframe>

Modern DevOps and backend engineering demand robust, flexible, and observable automation. Most workflow tools are either too generic, too complex, or not designed for Go developers. **iapetus** changes that: it's a fast, reliable, and developer-friendly workflow orchestrator written in Go, with first-class support for parallelism, observability, and testability.

---

## Why Choose iapetus?

- **Native Go API:** No YAML, no DSLs—just Go code.
- **DAG-Driven:** Express complex dependencies and parallelism with ease.
- **Observability Hooks:** Get deep insight into every task and workflow event.
- **Battle-Tested:** 100% test coverage, stress-tested, and used in real CI/CD pipelines.
- **Developer Joy:** Fluent APIs, custom assertions, and zero goroutine leaks.

---

## Getting Started

### Installation

```sh
go get github.com/yindia/iapetus
```

### Your First Workflow

```go
package main

import (
    "github.com/yindia/iapetus"
    "go.uber.org/zap"
    "time"
)

func main() {
    task := iapetus.NewTask("hello", 2*time.Second, nil).
        AddCommand("echo").
        AddArgs("Hello, world!").
        AssertOutputContains("Hello")

    workflow := iapetus.NewWorkflow("quickstart", zap.NewExample()).
        AddTask(*task)

    if err := workflow.Run(); err != nil {
        panic(err)
    }
}
```

---

## Features & Examples

### 1. Parallel, Dependency-Aware DAG Execution

Define tasks and their dependencies as a Directed Acyclic Graph (DAG). iapetus automatically schedules tasks in parallel, respecting dependencies.

```go
taskA := iapetus.NewTask("A", 2*time.Second, nil).AddCommand("echo").AddArgs("A")
taskB := iapetus.NewTask("B", 2*time.Second, nil).AddCommand("echo").AddArgs("B").DependsOn("A")
workflow := iapetus.NewWorkflow("dag", zap.NewNop()).AddTask(*taskA).AddTask(*taskB)
workflow.Run()
```

---

### 2. Custom & Built-in Assertions

Validate task outputs, exit codes, and more with built-in or custom assertions.

```go
task := &iapetus.Task{
    Name:    "verify",
    Command: "curl",
    Args:    []string{"-f", "http://localhost:8080"},
    Asserts: []func(*iapetus.Task) error{
        iapetus.AssertExitCode(0),
        iapetus.AssertOutputContains("Success"),
        func(t *iapetus.Task) error {
            if !strings.Contains(t.Actual.Output, "ready") {
                return fmt.Errorf("not ready")
            }
            return nil
        },
    },
}
```

---

### 3. Timeouts & Retries

Prevent hangs and handle flaky tasks gracefully.

```go
task := iapetus.NewTask("timeout", 1*time.Second, nil).
    AddCommand("sleep").
    AddArgs("10") // Will timeout after 1s

task.Retries = 3 // Retries up to 3 times on failure
```

---

### 4. Environment Variables

Set environment variables per task, either as a slice or a map.

```go
task := &iapetus.Task{
    Name:    "env-task",
    Command: "printenv",
    Args:    []string{"FOO"},
    Env:     []string{"FOO=bar"},
}
task.EnvMap = map[string]string{"FOO": "bar"}
```

---

### 5. Observability & Hooks

Get notified on every task event for deep observability and custom logic.

```go
workflow.AddOnTaskStartHook(func(task *iapetus.Task) { fmt.Println("Started:", task.Name) })
workflow.AddOnTaskSuccessHook(func(task *iapetus.Task) { fmt.Println("Success:", task.Name) })
workflow.AddOnTaskFailureHook(func(task *iapetus.Task, err error) { fmt.Println("Failed:", task.Name, err) })
workflow.AddOnTaskCompleteHook(func(task *iapetus.Task) { fmt.Println("Complete:", task.Name) })
```

---

### 6. PreRun & PostRun Hooks

Run setup and teardown logic at the workflow or task level.

```go
workflow.AddPreRunHook(func() error {
    fmt.Println("Workflow starting!")
    return nil
})
workflow.AddPostRunHook(func() error {
    fmt.Println("Workflow finished!")
    return nil
})
```

---

### 7. Flexible APIs

Choose between struct literals or fluent builder patterns for maximum flexibility.

```go
// Struct literal
task := &iapetus.Task{
    Name:    "literal",
    Command: "echo",
    Args:    []string{"hi"},
}

// Fluent builder
task := iapetus.NewTask("builder", 2*time.Second, nil).
    AddCommand("echo").
    AddArgs("hi")
```

---

### 8. DAG Validation

Detect cycles and missing dependencies before execution.

```go
task := &iapetus.Task{
    Name:    "bad",
    Depends: []string{"nonexistent"},
}
```

---

## Real-World Integrations

### Docker Example

```go
workflow := iapetus.NewWorkflow("docker", zap.NewNop())
workflow.AddTask(iapetus.Task{
    Name:    "Pull Alpine",
    Command: "docker",
    Args:    []string{"pull", "alpine"},
    Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)},
})
workflow.AddTask(iapetus.Task{
    Name:    "Run Echo",
    Command: "docker",
    Args:    []string{"run", "--rm", "alpine", "echo", "hello"},
    Asserts: []func(*iapetus.Task) error{iapetus.AssertOutputContains("hello")},
    Depends: []string{"Pull Alpine"},
})
workflow.Run()
```

---

### Kubernetes Example

```go
workflow := iapetus.NewWorkflow("k8s", zap.NewNop())
workflow.AddTask(iapetus.Task{
    Name:    "Create NS",
    Command: "kubectl",
    Args:    []string{"create", "ns", "test"},
    Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)},
})
workflow.AddTask(iapetus.Task{
    Name:    "Deploy Nginx",
    Command: "kubectl",
    Args:    []string{"create", "deployment", "nginx", "--image", "nginx", "-n", "test"},
    Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)},
    Depends: []string{"Create NS"},
})
workflow.Run()
```

---

## Reliability You Can Trust

- **100% Test Coverage:** Every line, every branch.
- **Stress & Concurrency Tested:** Zero goroutine leaks, even under heavy load.
- **Production Proven:** Used in real CI/CD, system test, and infrastructure automation pipelines.

---

## FAQ

**Q: What platforms does iapetus support?**  
A: Pure Go—works on Linux, macOS, Windows (Go 1.20+).

**Q: How do I add a custom assertion?**  
A: Write a `func(*iapetus.Task) error` and add it to your task.

**Q: Can I use iapetus for CI/CD, testing, or infra automation?**  
A: Absolutely! That's what it's built for.

---

## Get Involved

- **Star us on GitHub:** [iapetus](https://github.com/yindia/iapetus)
- **Open an issue:** Feature requests, bugs, and ideas welcome!
- **Contribute:** Fork, branch, PR—see the README for dev setup.

---

Happy automating!  
— Yuvraj & the iapetus team
