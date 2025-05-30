# iapetus Usage Guide

This guide covers the most common ways to use iapetus for workflow automation, testing, and orchestration.

---

## 🧩 Concepts

- **Workflow**: A collection of tasks with dependencies.
- **Task**: A single command or operation (e.g., `echo hello`).
- **Backend**: The environment where a task runs (e.g., bash, docker).
- **Assertion**: A check to validate task results (e.g., output contains "foo").
- **Hook**: Custom logic triggered on task events (start, success, failure).
- **Plugin**: A backend implementation that runs tasks in a specific environment (e.g., Docker, Bash, your own system).

---

## 🚦 Workflow Features & Fields

A **Workflow** orchestrates a set of tasks, manages dependencies, and provides global configuration.

### Workflow Fields

| Field    | Type                | Description |
|----------|---------------------|-------------|
| `Name`   | `string`            | Name of the workflow |
| `Steps`  | `[]*Task`           | List of tasks (steps) |
| `Backend`| `string`            | Default backend for all tasks (e.g., `bash`, `docker`) |
| `EnvMap` | `map[string]string` | Environment variables for all tasks |
| Hooks    | (various)           | Functions called on task events (start, success, failure, complete) |

#### Example

```go
workflow := iapetus.NewWorkflow("my-wf", zap.NewNop())
workflow.Backend = "bash" // Default backend for all tasks
workflow.EnvMap = map[string]string{"FOO": "bar"} // Global env vars
workflow.AddTask(*task1)
workflow.AddTask(*task2)
workflow.AddOnTaskSuccessHook(func(t *iapetus.Task) {
    fmt.Println("Task succeeded:", t.Name)
})
err := workflow.Run()
```

---

## 🧑‍💻 Task Features & Fields

A **Task** represents a single command or operation. Tasks can depend on each other, run in different environments, and have custom assertions.

### Task Fields

| Field      | Type                  | Description |
|------------|-----------------------|-------------|
| `Name`     | `string`              | Name of the task (unique in workflow) |
| `Command`  | `string`              | Command to execute (e.g., `echo`) |
| `Args`     | `[]string`            | Arguments for the command |
| `Timeout`  | `time.Duration`       | Max execution time (default: 30s, can override per-task) |
| `Retries`  | `int`                 | Number of retry attempts on failure |
| `Depends`  | `[]string`            | Names of tasks this task depends on |
| `EnvMap`   | `map[string]string`   | Environment variables for this task |
| `Image`    | `string`              | Container image (for containerized backends) |
| `Asserts`  | `[]func(*Task) error` | List of assertion functions |
| `Backend`  | `string`              | Backend for this task (overrides workflow default) |

#### Example

```go
task := iapetus.NewTask("hello", 5*time.Second, nil).
    AddCommand("echo").
    AddArgs("hello world").
    AssertOutputContains("hello")
task.Retries = 2 // Retry up to 2 times on failure
task.Timeout = 10 * time.Second // Custom timeout
task.EnvMap = map[string]string{"BAR": "baz"}
task.SetBackend("docker") // Run in Docker instead of default
```

#### Task Dependencies

```go
task2 := iapetus.NewTask("step2", 5*time.Second, nil).
    AddCommand("echo").
    AddArgs("second step")
task2.Depends = []string{"hello"} // Runs after "hello" task
```

---

## 🧪 Assertions

Assertions are checks that validate the result of a task. You can use built-in assertions or define your own.

### Built-in Assertions
- `AssertExitCode(int)`
- `AssertOutputEquals(string)`
- `AssertOutputContains(string)`
- `AssertOutputJsonEquals(string, ...string)`
- `AssertOutputMatchesRegexp(string)`

#### Example

```go
task.AssertExitCode(0)
task.AssertOutputContains("success")
```

### Custom Assertion Example

```go
task.AddAssertion(func(t *iapetus.Task) error {
    if !strings.Contains(t.Actual.Output, "success") {
        return fmt.Errorf("expected 'success' in output")
    }
    return nil
})
```

---

## 🔌 Plugin System: Backends

A **Backend** is a plugin that defines how tasks are executed (e.g., in bash, Docker, or your own system).

### Built-in Backends
- `bash` (default): Runs commands on the local shell
- `docker`: Runs commands in Docker containers

### How to Add a New Backend Plugin

1. **Implement the Backend interface:**

```go
type MyBackend struct{}

func (b *MyBackend) RunTask(task *iapetus.Task) error {
    // Your logic to run the task
    return nil
}

func (b *MyBackend) ValidateTask(task *iapetus.Task) error {
    // Optional: validate before running
    return nil
}
```

2. **Register your backend:**

```go
iapetus.RegisterBackend("my-backend", &MyBackend{})
```

3. **Use your backend in a workflow or task:**

- Set as the default for the workflow:

```go
workflow.Backend = "my-backend"
```

- Or override for a specific task:

```go
task.SetBackend("my-backend")
```

Now, when the workflow runs, tasks will be executed using your custom backend.

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

## 🔌 Using Hooks

Hooks let you run custom logic on task events:

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