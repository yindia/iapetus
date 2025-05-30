Tutorial: Getting Started with iapetus
======================================

This tutorial will walk you through installing iapetus, writing and running your first workflow in YAML and Go, using assertions and dependencies, running tasks in Docker, and extending iapetus with a custom backend.

Step 1: Installation
--------------------

1. **Install Go** (see https://golang.org/dl/)
2. **Clone the repo**

   .. code-block:: shell

      git clone https://github.com/yindia/iapetus.git
      cd iapetus
      go mod tidy

3. **(Optional) Use pixi for reproducible environments**

   .. code-block:: shell

      pixi shell

Step 2: Write Your First Workflow in YAML
-----------------------------------------

Create a file `workflow.yaml`:

.. code-block:: yaml

   name: hello-workflow
   backend: bash
   steps:
     - name: say-hello
       command: echo
       args: ["Hello, iapetus!"]
       raw_asserts:
         - output_contains: iapetus

Step 3: Run the Workflow
------------------------

Create a Go file (e.g., `main.go`):

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

You should see output showing the task running and passing.

Step 4: Write a Workflow in Go
------------------------------

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

Step 5: Add Assertions and Dependencies
---------------------------------------

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

Step 6: Use the Docker Backend
------------------------------

.. code-block:: go

   task := iapetus.NewTask("docker-echo", 5*time.Second, nil).
       AddCommand("echo").AddArgs("inside container").
       AssertOutputContains("inside container")
   task.SetBackend("docker")
   task.Image = "alpine:3.18"
   wf := iapetus.NewWorkflow("docker-demo", zap.NewNop())
   wf.AddTask(*task)
   wf.Run()

Step 7: Extend with a Custom Backend Plugin
-------------------------------------------

.. code-block:: go

   type MyBackend struct{}
   func (b *MyBackend) RunTask(task *iapetus.Task) error {
       // Custom logic
       return nil
   }
   func (b *MyBackend) ValidateTask(task *iapetus.Task) error { return nil }
   iapetus.RegisterBackend("my-backend", &MyBackend{})
   // Use in workflow or task as shown above

Debugging and Common Errors
---------------------------

- "command not found": Ensure the command exists in your environment or Docker image
- "permission denied": Check permissions
- "Go not installed": Install Go from https://golang.org/dl/
- For more, see the FAQ in the main README

Next Steps
----------

- See :doc:`usage` for more examples and best practices
- See :doc:`api` for full API and YAML reference
- Explore the [GitHub repo](https://github.com/yindia/iapetus) for more examples 