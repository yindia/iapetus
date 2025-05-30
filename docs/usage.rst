Usage Guide
===========

What is iapetus?
----------------

**iapetus** is a workflow engine for automating, testing, and orchestrating command-line tasks. It is ideal for CI/CD, DevOps, E2E testing, and any scenario where you want to run and validate CLI tools in sequence or parallel, with dependencies and assertions.

When to use iapetus:
- Automate complex shell or container workflows
- Validate CLI tools, microservices, or scripts
- Run integration/E2E tests for your stack
- Orchestrate tasks in CI/CD pipelines

Key Concepts
------------

- **Workflow**: A collection of tasks with dependencies and global config.
- **Task**: A single command or operation, with its own config and assertions.
- **Backend**: The environment where a task runs (e.g., bash, docker, custom plugin).
- **Assertion**: A check to validate task results (e.g., output contains "foo").
- **Plugin**: A backend implementation that runs tasks in a specific environment.
- **Hook**: Custom logic triggered on task events (start, success, failure).

YAML vs Go Usage
----------------

You can define workflows in YAML (for config-driven use) or Go (for full flexibility).

YAML Example:
^^^^^^^^^^^^^

.. code-block:: yaml

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

To run a YAML workflow:

.. code-block:: go

   wf, err := iapetus.LoadWorkflowFromYAML("workflow.yaml")
   if err != nil { log.Fatal(err) }
   err = wf.Run()

Go Example:
^^^^^^^^^^^

.. code-block:: go

   task := iapetus.NewTask("hello", 5*time.Second, nil).
       AddCommand("echo").
       AddArgs("hello world").
       AssertOutputContains("hello")
   workflow := iapetus.NewWorkflow("my-wf", zap.NewNop())
   workflow.Backend = "bash"
   workflow.AddTask(*task)
   err := workflow.Run()

Defining and Running Workflows
------------------------------

1. **Define tasks** (with commands, args, env, assertions, etc.)
2. **Add tasks to a workflow** (set backend, env, hooks, etc.)
3. **Run the workflow**

.. code-block:: go

   task1 := iapetus.NewTask("step1", 5*time.Second, nil).
       AddCommand("echo").AddArgs("foo").AssertOutputContains("foo")
   task2 := iapetus.NewTask("step2", 5*time.Second, nil).
       AddCommand("echo").AddArgs("bar")
   task2.Depends = []string{"step1"}
   wf := iapetus.NewWorkflow("demo", zap.NewNop())
   wf.AddTask(*task1)
   wf.AddTask(*task2)
   wf.Run()

Best Practices
--------------

- Use `Retries` for flaky steps
- Use `Timeout` to prevent hangs
- Use `EnvMap` for environment variables
- Use hooks for custom logging or metrics
- Keep YAML workflows simple for CI/CD

Troubleshooting
---------------

- "command not found": Ensure the command exists in your environment or Docker image
- "permission denied": Check permissions
- "Go not installed": Install Go from https://golang.org/dl/
- For more, see the FAQ in the main README

See Also
--------

- :doc:`tutorial`
- :doc:`api`
- `Full GoDoc <https://pkg.go.dev/github.com/yindia/iapetus>`_ 