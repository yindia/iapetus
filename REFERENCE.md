# iapetus API & YAML Reference

This document provides a detailed technical reference for the iapetus Go API and YAML schema.

---

## 🏗️ Go API Reference

### Workflow

```go
type Workflow struct {
    Name    string
    Steps   []*Task
    Backend string // default backend for all tasks
    EnvMap  map[string]string
    // ... hooks, logger, etc.
}
```

### Task

```go
type Task struct {
    Name     string
    Command  string
    Args     []string
    Timeout  time.Duration
    Retries  int
    Depends  []string
    EnvMap   map[string]string
    Image    string // for container backends
    Asserts  []func(*Task) error
    Backend  string // optional override
}
```

### Backend Interface

```go
type Backend interface {
    RunTask(task *Task) error
    ValidateTask(task *Task) error
}
```

Register a backend:

```go
iapetus.RegisterBackend("my-backend", &MyBackend{})
```

### Assertion Functions

- `AssertExitCode(expected int)`
- `AssertOutputEquals(expected string)`
- `AssertOutputContains(substr string)`
- `AssertOutputJsonEquals(expected string, skipJsonNodes ...string)`
- `AssertOutputMatchesRegexp(pattern string)`

Custom assertion:

```go
func AssertMyCheck(expected string) func(*iapetus.Task) error {
    return func(t *iapetus.Task) error {
        if t.Actual.Output != expected {
            return fmt.Errorf("want %q, got %q", expected, t.Actual.Output)
        }
        return nil
    }
}
```

---

## 📝 YAML Schema Reference

### Workflow YAML

```yaml
name: my-wf
backend: bash # default backend for all steps
env_map:
  FOO: bar
steps:
  - name: hello
    command: echo
    args: ["hello"]
    timeout: 5s
    backend: docker # optional, overrides workflow backend
    env_map:
      BAR: baz
    raw_asserts:
      - output_contains: hello
```

### Supported Assertion Types in YAML

- `exit_code: 0`
- `output_equals: "foo"`
- `output_contains: "bar"`
- `output_json_equals: '{"foo": 1}'`
- `output_matches_regexp: '^foo.*$'`
- `skip_json_nodes: ["foo.bar"]` (for JSON assertions)

---

## 🛠️ Extending iapetus

### Add a Custom Backend

```go
type MyBackend struct{}
func (b *MyBackend) RunTask(task *iapetus.Task) error { ... }
func (b *MyBackend) ValidateTask(task *iapetus.Task) error { return nil }
iapetus.RegisterBackend("my-backend", &MyBackend{})
```

### Add a Custom Assertion

```go
func AssertMyCheck(expected string) func(*iapetus.Task) error {
    return func(t *iapetus.Task) error {
        if t.Actual.Output != expected {
            return fmt.Errorf("want %q, got %q", expected, t.Actual.Output)
        }
        return nil
    }
}
```

---

## [Back to README](README.md) 