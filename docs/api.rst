API Reference
============

.. raw:: html

   <hr style="margin-top: 0; margin-bottom: 1.5em; border: none; border-top: 2px solid #eee;"/>

This page documents the main Go API for iapetus. It covers the core structs, interfaces, and extension points. For YAML usage, see the YAML Reference.

Workflow Struct 🗂️
------------------

The `Workflow` struct defines a collection of tasks, their dependencies, and global configuration. It is the top-level object for running a workflow.

.. code-block:: go

   type Workflow struct {
       Name    string              // Name of the workflow
       Steps   []*Task             // List of tasks (steps)
       Backend string              // Default backend for all tasks (e.g., "bash", "docker")
       EnvMap  map[string]string   // Environment variables for all tasks
       // ... hooks, logger, etc.
   }

**Fields:**
- `Name`: Human-readable workflow name.
- `Steps`: List of tasks (see below).
- `Backend`: Default backend for all steps (can be overridden per-task).
- `EnvMap`: Environment variables for all steps (can be overridden per-task).
- Hooks, logger, and other advanced fields are available for extensibility.

.. admonition:: Example
   :class: tip

   .. code-block:: go

      wf := iapetus.NewWorkflow("my-wf", zap.NewNop())
      wf.Backend = "bash"
      wf.EnvMap = map[string]string{"FOO": "bar"}
      wf.AddTask(*task)
      wf.Run()

Task Struct 🏃
-------------

A `Task` represents a single step in a workflow. Each task can have its own command, arguments, environment, timeout, assertions, and backend.

.. code-block:: go

   type Task struct {
       Name     string                // Name of the task (unique in workflow)
       Command  string                // Command to execute (e.g., "echo")
       Args     []string              // Arguments for the command
       Timeout  time.Duration         // Max execution time (default: 30s, can override per-task)
       Retries  int                   // Number of retry attempts on failure
       Depends  []string              // Names of tasks this task depends on
       EnvMap   map[string]string     // Environment variables for this task
       Image    string                // Container image (for containerized backends)
       Asserts  []func(*Task) error   // List of assertion functions
       Backend  string                // Backend for this task (overrides workflow default)
   }

**Fields:**
- `Name`: Unique name for the task.
- `Command`: The executable or shell command.
- `Args`: Arguments to pass to the command.
- `Timeout`: Maximum allowed execution time (default: 30s).
- `Retries`: Number of times to retry on failure.
- `Depends`: List of task names this task depends on (for DAG execution).
- `EnvMap`: Environment variables for this task (overrides workflow-level vars).
- `Image`: Docker image (for Docker backend).
- `Asserts`: List of assertion functions (see below).
- `Backend`: Backend to use for this task (overrides workflow default).

.. admonition:: Example
   :class: tip

   .. code-block:: go

      task := iapetus.NewTask("hello", 5*time.Second, nil).
          AddCommand("echo").
          AddArgs("Hello, world!").
          AssertOutputContains("Hello")
      task.Timeout = 10 * time.Second
      task.Retries = 2
      task.EnvMap = map[string]string{"FOO": "bar"}
      task.Backend = "docker"
      task.Image = "alpine:3.18"

Backend Interface & Plugins 🔌
-----------------------------

The `Backend` interface allows you to add new ways to run tasks (e.g., Docker, Kubernetes, SSH). Built-in backends include Bash and Docker.

.. code-block:: go

   type Backend interface {
       RunTask(task *Task) error
       ValidateTask(task *Task) error
       GetName() string
       GetStatus() string
   }

- `RunTask`: Executes the given task and populates its result fields.
- `ValidateTask`: Checks if the task is valid for this backend (e.g., required fields).
- `GetName`: Returns the backend's name (for registry and diagnostics).
- `GetStatus`: Returns a status string (e.g., "available", "unavailable").

.. admonition:: Registering a Backend
   :class: tip

   .. code-block:: go

      iapetus.RegisterBackend("my-backend", &MyBackend{})

.. admonition:: Example Plugin
   :class: tip

   .. code-block:: go

      type MyBackend struct{}
      func (b *MyBackend) RunTask(task *iapetus.Task) error { /* ... */ }
      func (b *MyBackend) ValidateTask(task *iapetus.Task) error { return nil }
      func (b *MyBackend) GetName() string { return "my-backend" }
      func (b *MyBackend) GetStatus() string { return "available" }

Assertion Functions ✅
---------------------

Assertions are checks that validate the result of a task. You can use built-in assertions or write your own.

**Built-in assertions:**

- `AssertExitCode(expected int)`
- `AssertOutputEquals(expected string)`
- `AssertOutputContains(substr string)`
- `AssertOutputJsonEquals(expected string, skipJsonNodes ...string)`
- `AssertOutputMatchesRegexp(pattern string)`

.. admonition:: Custom assertion example
   :class: tip

   .. code-block:: go

      func AssertMyCheck(expected string) func(*iapetus.Task) error {
          return func(t *iapetus.Task) error {
              if t.Actual.Output != expected {
                  return fmt.Errorf("want %q, got %q", expected, t.Actual.Output)
              }
              return nil
          }
      }

Hooks 🪝
-------

Hooks let you run custom logic on task events. Use hooks for logging, metrics, or notifications.

- `AddOnTaskStartHook(func(*Task))`
- `AddOnTaskSuccessHook(func(*Task))`
- `AddOnTaskFailureHook(func(*Task, error))`
- `AddOnTaskCompleteHook(func(*Task))`

YAML Schema Reference 📄
-----------------------

For YAML usage, see the YAML Reference. Example:

.. code-block:: yaml

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
       image: alpine:3.18
       env_map:
         BAR: baz
       raw_asserts:
         - output_contains: hello

Supported assertion types in YAML:

- `exit_code: 0`
- `output_equals: "foo"`
- `output_contains: "bar"`
- `output_json_equals: '{"foo": 1}'`
- `output_matches_regexp: '^foo.*$'`
- `skip_json_nodes: ["foo.bar"]` (for JSON assertions)

For more, see the `GoDoc <https://pkg.go.dev/github.com/yindia/iapetus>`_. 

.. raw:: html

   <hr style="margin-top: 1.5em; margin-bottom: 0; border: none; border-top: 2px solid #eee;"/> 