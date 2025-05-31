Concepts
========

.. raw:: html

   <hr style="margin-top: 0; margin-bottom: 1.5em; border: none; border-top: 2px solid #eee;"/>

📖 **iapetus is like a recipe book for automating command-line tasks. Here are the core concepts:**

- **Workflow** 🗂️: A workflow is a collection of tasks, their dependencies, and global configuration. Think of it as a recipe that describes the order and logic of your automation.

  .. code-block:: yaml

     name: my-workflow
     steps:
       - name: build
         command: make
         args: [build]
       - name: test
         command: make
         args: [test]
         depends: [build]

- **Task** 🏃: A single command or step in your workflow. Each task can have its own command, arguments, environment, timeout, assertions, and dependencies.

  .. code-block:: go

     task := iapetus.NewTask("hello", 5*time.Second, nil).
         AddCommand("echo").
         AddArgs("Hello, world!")

- **Backend** 🖥️: The environment where a task runs. Built-in backends include Bash (local shell) and Docker (container). You can add your own (e.g., Kubernetes, SSH).

  .. code-block:: yaml

     backend: docker
     steps:
       - name: run-in-container
         image: alpine:3.18
         command: echo
         args: ["hello from docker"]

- **Assertion** ✅: A check to validate task results, such as output, exit code, or JSON. Assertions help ensure your workflow behaves as expected.

  .. code-block:: yaml

     raw_asserts:
       - output_contains: success
       - exit_code: 0

- **Plugin** 🔌: Extend iapetus by adding new backends or assertion types. Plugins are Go code that implement the Backend interface and are registered at startup.

  .. code-block:: go

     type MyBackend struct{}
     func (b *MyBackend) RunTask(task *iapetus.Task) error { /* ... */ }
     func (b *MyBackend) ValidateTask(task *iapetus.Task) error { return nil }
     iapetus.RegisterBackend("my-backend", &MyBackend{})

- **Hook** 🪝: Custom logic that runs on task events (start, success, failure, complete). Use hooks for logging, metrics, or notifications.

  .. code-block:: go

     workflow.AddOnTaskSuccessHook(func(t *iapetus.Task) {
         fmt.Println("Task succeeded:", t.Name)
     })

.. raw:: html

   <hr style="margin-top: 1.5em; margin-bottom: 0; border: none; border-top: 2px solid #eee;"/>

**See also:** :doc:`api`, :doc:`yaml` 