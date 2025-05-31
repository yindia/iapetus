How-To Recipes
==============

.. raw:: html

   <hr style="margin-top: 0; margin-bottom: 1.5em; border: none; border-top: 2px solid #eee;"/>

This section answers common questions and shows practical patterns for using iapetus. Each recipe includes a short explanation and example.

How do I run a shell command and check output? 📝
------------------------------------------------
To run a shell command and assert on its output, define a step with the command, arguments, and an assertion:

.. code-block:: yaml

   steps:
     - name: check-echo
       command: echo
       args: ["hello"]
       raw_asserts:
         - output_contains: hello

- `raw_asserts` lets you check output, exit code, or other conditions.
- You can use multiple assertions per step.

How do I run steps in parallel? ⚡️
----------------------------------
By default, steps without dependencies run in parallel. Just define them without `depends`:

.. code-block:: yaml

   steps:
     - name: step1
       command: echo
       args: ["foo"]
     - name: step2
       command: echo
       args: ["bar"]

- If you want sequential execution, use `depends` to specify dependencies.

How do I pass environment variables? 🌱
--------------------------------------
Use `env_map` in YAML or `EnvMap` in Go to set environment variables for a step or the whole workflow.

.. code-block:: yaml

   env_map:
     FOO: bar
   steps:
     - name: print-env
       command: printenv
       args: ["FOO"]

- Per-step `env_map` overrides workflow-level variables.

How do I retry on failure? 🔁
----------------------------
Set `retries` in YAML or `task.Retries` in Go to automatically retry a step if it fails.

.. code-block:: yaml

   steps:
     - name: flaky-step
       command: ./sometimes-fails.sh
       retries: 2

.. code-block:: go

   task.Retries = 2

- Retries are useful for flaky network or integration steps.

How do I use Docker for isolation? 🐳
------------------------------------
Run steps in containers by setting `backend: docker` and specifying an `image`.

.. code-block:: yaml

   steps:
     - name: run-in-docker
       backend: docker
       image: alpine:3.18
       command: echo
       args: ["hello from docker"]

- Use Docker to match production environments or isolate dependencies.
- You can set the default backend for all steps at the workflow level.

How do I add a custom check/assertion? 🧪
----------------------------------------
You can add custom assertions in Go for advanced checks:

.. code-block:: go

   task.AddAssertion(func(t *iapetus.Task) error {
       if !strings.Contains(t.Actual.Output, "success") {
           return fmt.Errorf("expected 'success' in output")
       }
       return nil
   })

- Custom assertions let you check any property of the task result.
- You can combine built-in and custom assertions.

How do I debug a failing workflow? 🐞
------------------------------------
- Check the output and error logs for each step (see `t.Actual.Output` and `t.Actual.Error`).
- Use hooks to add custom logging or metrics (see the Concepts and API docs).
- Run with increased logging or in a local shell for easier troubleshooting.
- See :doc:`faq` for more tips and common issues.

.. raw:: html

   <hr style="margin-top: 1.5em; margin-bottom: 0; border: none; border-top: 2px solid #eee;"/> 