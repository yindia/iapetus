API Reference
============

Workflow Struct
---------------

.. code-block:: go

   type Workflow struct {
       Name    string              // Name of the workflow
       Steps   []*Task             // List of tasks (steps)
       Backend string              // Default backend for all tasks (e.g., "bash", "docker")
       EnvMap  map[string]string   // Environment variables for all tasks
       // ... hooks, logger, etc.
   }

Task Struct
-----------

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

Backend Interface & Plugins
--------------------------

.. code-block:: go

   type Backend interface {
       RunTask(task *Task) error
       ValidateTask(task *Task) error
   }

To register a backend:

.. code-block:: go

   iapetus.RegisterBackend("my-backend", &MyBackend{})

Assertion Functions
-------------------

Built-in assertions:

- AssertExitCode(expected int)
- AssertOutputEquals(expected string)
- AssertOutputContains(substr string)
- AssertOutputJsonEquals(expected string, skipJsonNodes ...string)
- AssertOutputMatchesRegexp(pattern string)

Custom assertion example:

.. code-block:: go

   func AssertMyCheck(expected string) func(*iapetus.Task) error {
       return func(t *iapetus.Task) error {
           if t.Actual.Output != expected {
               return fmt.Errorf("want %q, got %q", expected, t.Actual.Output)
           }
           return nil
       }
   }

Hooks
-----

Hooks let you run custom logic on task events:

- AddOnTaskStartHook(func(*Task))
- AddOnTaskSuccessHook(func(*Task))
- AddOnTaskFailureHook(func(*Task, error))
- AddOnTaskCompleteHook(func(*Task))

YAML Schema Reference
---------------------

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
       env_map:
         BAR: baz
       raw_asserts:
         - output_contains: hello

Supported assertion types in YAML:

- exit_code: 0
- output_equals: "foo"
- output_contains: "bar"
- output_json_equals: '{"foo": 1}'
- output_matches_regexp: '^foo.*$'
- skip_json_nodes: ["foo.bar"] (for JSON assertions)

For more, see the `GoDoc <https://pkg.go.dev/github.com/yindia/iapetus>`_. 