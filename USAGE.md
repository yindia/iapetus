# iapetus Usage Guide

This guide covers the most common ways to use iapetus for workflow automation, testing, and orchestration.

---

## 🧩 Concepts

- **Workflow**: A collection of tasks with dependencies.
- **Task**: A single command or operation (e.g., `echo hello`).
- **Backend**: The environment where a task runs (e.g., bash, docker).
- **Assertion**: A check to validate task results (e.g., output contains "foo").
- **Hook**: Custom logic triggered on task events (start, success, failure).

---

## 🚦 Creating a Workflow in Go

```go
import (
    "github.com/yindia/iapetus"
    "go.uber.org/zap"
    "time"
)

task := iapetus.NewTask("hello", 5*time.Second, nil).
    AddCommand("echo").
    AddArgs("hello world").
    AssertOutputContains("hello")
workflow := iapetus.NewWorkflow("my-wf", zap.NewNop())
workflow.Backend = "bash"
workflow.AddTask(*task)
err := workflow.Run()
```

---

## 📄 Creating a Workflow in YAML

```yaml
name: my-wf
backend: bash
env_map:
  FOO: bar
steps:
  - name: hello
    command: echo
    args: ["hello"]
    timeout: 5s
    backend: docker
    env_map:
      BAR: baz
    raw_asserts:
      - output_contains: hello
```

Load and run in Go:

```go
wf, err := iapetus.LoadWorkflowFromYAML("workflow.yaml")
if err != nil { log.Fatal(err) }
err = wf.Run()
```

---

## 🧪 Using Assertions

Built-in assertions:
- `AssertExitCode(int)`
- `AssertOutputEquals(string)`
- `AssertOutputContains(string)`
- `AssertOutputJsonEquals(string, ...string)`
- `AssertOutputMatchesRegexp(string)`

Custom assertion example:

```go
task.AddAssertion(func(t *iapetus.Task) error {
    if !strings.Contains(t.Actual.Output, "success") {
        return fmt.Errorf("expected 'success' in output")
    }
    return nil
})
```

---

## 🔌 Using Hooks

```go
workflow.AddOnTaskStartHook(func(t *iapetus.Task) {
    fmt.Println("Starting:", t.Name)
})
workflow.AddOnTaskSuccessHook(func(t *iapetus.Task) {
    fmt.Println("Success:", t.Name)
})
workflow.AddOnTaskFailureHook(func(t *iapetus.Task, err error) {
    fmt.Println("Failed:", t.Name, err)
})
```

---

## 🐳 Running in Docker

Set `workflow.Backend = "docker"` or `task.SetBackend("docker")` to run steps in containers. Use the `Image` field to specify the container image.

---

## 🛠️ Best Practices

- Use `Retries` for flaky steps
- Use `Timeout` to prevent hangs
- Use `EnvMap` for environment variables
- Use hooks for custom logging or metrics
- Keep YAML workflows simple for CI/CD

---

## 🐞 Troubleshooting

- "command not found": Ensure the command exists in your environment or Docker image
- "permission denied": Check permissions
- "Go not installed": Install Go from [golang.org](https://golang.org/dl/)
- For more, see [FAQ](README.md#faq) 