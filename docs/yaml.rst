YAML Reference
==============

iapetus workflows can be defined in YAML for easy, no-code automation. This page explains the YAML schema, field meanings, and best practices.

YAML Schema
-----------

A typical workflow YAML file looks like this:

.. code-block:: yaml

   name: my-wf                # (required) Name of the workflow
   backend: bash              # (optional) Default backend for all steps ("bash", "docker", or custom)
   env_map:                   # (optional) Environment variables for all steps
     FOO: bar
   steps:
     - name: hello            # (required) Name of the step (unique)
       command: echo          # (required) Command to run
       args: ["hello"]        # (optional) Arguments for the command
       timeout: 5s            # (optional) Max execution time (e.g., 5s, 1m)
       backend: docker        # (optional) Backend for this step (overrides workflow backend)
       image: alpine:3.18     # (required for docker) Docker image to use
       env_map:               # (optional) Env vars for this step (overrides workflow env_map)
         BAR: baz
       retries: 2             # (optional) Number of retry attempts on failure
       depends: [other-step]  # (optional) List of step names this step depends on
       raw_asserts:           # (optional) List of assertions to check after execution
         - output_contains: hello
         - exit_code: 0

**Field Explanations:**
- `name`: Unique name for the workflow or step.
- `backend`: Where to run steps ("bash" for local shell, "docker" for containers, or a custom backend).
- `env_map`: Key-value pairs of environment variables. Can be set globally or per-step.
- `steps`: List of steps (tasks) to run.
- `command`: The executable or shell command to run.
- `args`: List of arguments for the command.
- `timeout`: Maximum allowed time for the step (e.g., 10s, 2m). Default is 30s.
- `image`: Docker image to use (required for Docker backend).
- `retries`: Number of times to retry the step on failure.
- `depends`: List of step names this step depends on (for ordering and parallelism).
- `raw_asserts`: List of assertions to check after the step runs.

**Tips:**
- Indent with spaces, not tabs.
- Use quotes for strings with special characters.
- You can override the backend, env, and timeout per step.
- Steps without `depends` run in parallel.

Supported assertion types:
-------------------------
- `exit_code: 0` — Check the exit code of the command.
- `output_equals: "foo"` — Output must exactly match the string.
- `output_contains: "bar"` — Output must contain the substring.
- `output_json_equals: '{"foo": 1}'` — Output must match the given JSON.
- `output_matches_regexp: '^foo.*$'` — Output must match the regular expression.
- `skip_json_nodes: ["foo.bar"]` — Used with JSON assertions to ignore certain fields.

Backend options:
----------------
- `bash`: Runs the command in your local shell (default, works everywhere).
- `docker`: Runs the command in a Docker container (requires `image`).
- Custom: You can register your own backend in Go and reference it by name.

Example: Minimal Workflow
------------------------

.. code-block:: yaml

   name: minimal
   steps:
     - name: hello
       command: echo
       args: ["Hello, world!"]
       raw_asserts:
         - output_contains: Hello

Example: Docker Workflow
-----------------------

.. code-block:: yaml

   name: docker-demo
   backend: docker
   steps:
     - name: run-in-container
       image: alpine:3.18
       command: echo
       args: ["hello from docker"]
       raw_asserts:
         - output_contains: hello

See also: :doc:`api` for Go API details and advanced usage. 