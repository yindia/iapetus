Tutorial: Getting Started with iapetus
======================================

This tutorial will guide you through installing iapetus, writing and running your first workflow in YAML and Go, using assertions and dependencies, running tasks in Docker, and extending iapetus with a custom backend. Each step explains not just what to do, but why.

Step 1: Installation
--------------------

Before you can use iapetus, you need Go installed on your system. Go is required to build and run iapetus workflows.

1. **Install Go** ([download here](https://golang.org/dl/))
   - Verify installation:

     .. code-block:: shell

        go version

2. **Clone the iapetus repository**
   - This gives you access to all examples and the latest code.

     .. code-block:: shell

        git clone https://github.com/yindia/iapetus.git
        cd iapetus
        go mod tidy  # Download Go dependencies

3. **(Optional) Use pixi for reproducible environments**
   - [pixi](https://pixi.sh/) helps you manage dependencies and environments, especially for collaborative or CI/CD setups.

     .. code-block:: shell

        pixi shell

Step 2: Write Your First Workflow in YAML
-----------------------------------------

YAML is a simple way to define workflows without writing Go code. Create a file called `workflow.yaml`:

.. code-block:: yaml

   name: hello-workflow
   backend: bash
   steps:
     - name: say-hello
       command: echo
       args: ["Hello, iapetus!"]
       raw_asserts:
         - output_contains: iapetus

**Explanation:**
- `name`: The workflow's name.
- `backend`: Where tasks run ("bash" = local shell).
- `steps`: List of tasks. Each task has a name, command, arguments, and assertions.
- `raw_asserts`: Checks that the output contains "iapetus" (verifies success).

Step 3: Run the Workflow
------------------------

To run a YAML workflow, you need a small Go program to load and execute it. Create `main.go`:

.. code-block:: go

   package main
   import (
       "github.com/yindia/iapetus"
       "log"
   )
   func main() {
       wf, err := iapetus.LoadWorkflowFromYAML("workflow.yaml")
       if err != nil { log.Fatal(err) }
       if err := wf.Run(); err != nil {
           log.Fatalf("Workflow failed: %v", err)
       }
   }

Run it:

.. code-block:: shell

   go run main.go

**What happens:**
- iapetus loads your YAML, runs each step, and prints results.
- If the assertion fails, you'll see an error.

Step 4: Write a Workflow in Go
------------------------------

For more flexibility, you can define workflows in Go. This is useful for dynamic logic, custom hooks, or programmatic task generation.

.. code-block:: go

   import (
       "github.com/yindia/iapetus"
       "go.uber.org/zap"
       "time"
   )
   task := iapetus.NewTask("hello", 5*time.Second, nil).
       AddCommand("echo").
       AddArgs("Hello, iapetus!").
       AssertOutputContains("iapetus")
   workflow := iapetus.NewWorkflow("my-wf", zap.NewNop())
   workflow.AddTask(*task)
   workflow.Run()

**Why use Go?**
- Add custom logic, hooks, or assertions.
- Integrate with other Go code or libraries.

Step 5: Add Assertions and Dependencies
---------------------------------------

Assertions check that your tasks did what you expected. Dependencies let you control execution order.

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

**Explanation:**
- `AssertOutputContains("foo")`: Checks output of step1.
- `task2.Depends = ["step1"]`: step2 runs only after step1 succeeds.

Step 6: Use the Docker Backend
------------------------------

You can run tasks inside Docker containers for isolation or to match production environments.

.. code-block:: go

   task := iapetus.NewTask("docker-echo", 5*time.Second, nil).
       AddCommand("echo").AddArgs("inside container").
       AssertOutputContains("inside container")
   task.SetBackend("docker")
   task.Image = "alpine:3.18"
   wf := iapetus.NewWorkflow("docker-demo", zap.NewNop())
   wf.AddTask(*task)
   wf.Run()

**Tips:**
- Set `task.Image` to the Docker image you want.
- Use Docker backend for clean, reproducible builds/tests.

Step 7: Extend with a Custom Backend Plugin
-------------------------------------------

You can add your own backend (e.g., run tasks on Kubernetes, SSH, etc) by implementing the Backend interface.

.. code-block:: go

   type MyBackend struct{}
   func (b *MyBackend) RunTask(task *iapetus.Task) error {
       // Custom logic
       return nil
   }
   func (b *MyBackend) ValidateTask(task *iapetus.Task) error { return nil }
   iapetus.RegisterBackend("my-backend", &MyBackend{})
   // Use in workflow or task as shown above

**Why plugins?**
- Integrate with any environment or system.
- Add new ways to run or validate tasks.

Debugging and Common Errors
---------------------------

- "command not found": Ensure the command exists in your environment or Docker image.
- "permission denied": Check file and Docker permissions.
- "Go not installed": Install Go from https://golang.org/dl/
- For more, see the FAQ in the main README or docs.

Next Steps
----------

- See :doc:`usage` for more examples and best practices
- See :doc:`api` for full API and YAML reference
- Explore the [GitHub repo](https://github.com/yindia/iapetus) for more examples 