Tutorials
=========

Step-by-step guides for all skill levels.

Tutorial 1: Run Your First Workflow (YAML)
------------------------------------------
.. code-block:: yaml

   name: hello-workflow
   backend: bash
   steps:
     - name: say-hello
       command: echo
       args: ["Hello, iapetus!"]
       raw_asserts:
         - output_contains: iapetus

Run it with:
.. code-block:: shell

   go run main_docker.go

Tutorial 2: Add Checks/Assertions
---------------------------------
Change `raw_asserts` to check for different output, or add more assertions.

Tutorial 3: Add Dependencies
----------------------------
Add `depends: [step1]` to a step to make it wait for another.

Tutorial 4: Run in Docker
-------------------------
Set `backend: docker` and `image: alpine:3.18` in your YAML.

Tutorial 5: Write a Workflow in Go
----------------------------------
.. code-block:: go

   task := iapetus.NewTask("hello", 5*time.Second, nil).
       AddCommand("echo").
       AddArgs("Hello, iapetus!").
       AssertOutputContains("iapetus")
   workflow := iapetus.NewWorkflow("my-wf", zap.NewNop())
   workflow.AddTask(*task)
   workflow.Run()

Tutorial 6: Add Your Own Plugin (Advanced)
------------------------------------------
See :doc:`api` for the Backend interface and plugin registration.

**Next:** :doc:`howto` 