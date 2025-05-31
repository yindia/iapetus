FAQ
===

.. raw:: html

   <hr style="margin-top: 0; margin-bottom: 1.5em; border: none; border-top: 2px solid #eee;"/>

Why use iapetus over X? 🤔
-------------------------
- **YAML and Go support**: Define workflows as code or config.
- **Plugin system**: Add any backend (Docker, Kubernetes, SSH, etc).
- **Easy assertions and dependencies**: Built-in checks and DAG execution.
- **Open source, no vendor lock-in**.

Can I use this in CI/CD? 🚦
--------------------------
Yes! iapetus is designed for automation and CI/CD. It runs anywhere Go runs, and integrates easily with GitHub Actions, GitLab CI, Jenkins, and more.

How do I contribute? 🤝
----------------------
- See :doc:`contributing` for setup, code style, and PR guidelines.
- All contributions—code, docs, tests, issues—are welcome!

Where to get help? 🆘
---------------------
- GitHub Issues (for bugs and feature requests)
- Discussions (for questions, ideas, and community support)
- This documentation (for usage, API, and troubleshooting)

Common errors and solutions 🛠️
-----------------------------
.. admonition:: Troubleshooting
   :class: tip

   - **"command not found"**: Check your command or Docker image. Make sure the binary exists in the container or shell.
   - **"permission denied"**: Check file permissions and Docker access. Try running with elevated privileges if needed.
   - **"Go not installed"**: Install Go from https://golang.org/dl/ and ensure it's in your PATH.
   - **Docker errors**: Make sure Docker is installed, running, and your user has permission to run containers. See `iapetus.GetStatus()` for backend availability.
   - **YAML parse errors**: Check indentation, use spaces not tabs, and quote strings with special characters.
   - **Assertion failed**: Review the step's output and assertion logic. Use `output_contains` or custom assertions for flexible checks.
   - **Timeouts**: Increase the `timeout` field if your step needs more time.
   - **Retries not working**: Ensure `retries` is set at the step/task level.

How do I debug a failing workflow? 🐞
------------------------------------
- Check the output and error logs for each step.
- Use hooks for custom logging or metrics.
- Run with increased logging or in a local shell for easier troubleshooting.
- See :doc:`howto` for debugging tips.

How do I run steps in Docker? 🐳
------------------------------
- Set `backend: docker` and specify an `image` for your step or workflow.
- Make sure Docker is installed and running on your system.
- See the YAML Reference for examples.

How do I add a custom backend or assertion? 🔌
---------------------------------------------
- Implement the `Backend` interface in Go and register it with `iapetus.RegisterBackend()`.
- Add custom assertion functions in Go and attach them to your tasks.
- See the API Reference for plugin and assertion examples.

Where can I find more examples? 📚
----------------------------------
- See the :doc:`tutorial`, :doc:`howto`, and :doc:`yaml` pages.
- Explore the [GitHub repo](https://github.com/yindia/iapetus) for real-world workflows and code.

.. raw:: html

   <hr style="margin-top: 1.5em; margin-bottom: 0; border: none; border-top: 2px solid #eee;"/> 