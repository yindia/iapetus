How-To Recipes
==============

How do I run a shell command and check output?
----------------------------------------------
.. code-block:: yaml

   steps:
     - name: check-echo
       command: echo
       args: ["hello"]
       raw_asserts:
         - output_contains: hello

How do I run steps in parallel?
-------------------------------
Just define multiple steps without dependencies.

How do I pass environment variables?
------------------------------------
Use `env_map` in YAML or `EnvMap` in Go.

How do I retry on failure?
--------------------------
Set `retries: 2` in YAML or `task.Retries = 2` in Go.

How do I use Docker for isolation?
----------------------------------
Set `backend: docker` and `image: alpine:3.18` for your step.

How do I add a custom check/assertion?
--------------------------------------
.. code-block:: go

   task.AddAssertion(func(t *iapetus.Task) error {
       if !strings.Contains(t.Actual.Output, "success") {
           return fmt.Errorf("expected 'success' in output")
       }
       return nil
   })

How do I debug a failing workflow?
----------------------------------
- Check the output and error logs
- Use hooks for custom logging
- See :doc:`faq` 