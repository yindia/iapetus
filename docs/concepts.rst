Concepts
========

iapetus is like a recipe book for automating command-line tasks.

- **Workflow**: Like a recipe—a list of steps (tasks) to run.
- **Task**: A single command (e.g., `echo hello`).
- **Backend**: Where the task runs (your shell, Docker, or your own system).
- **Assertion**: How you check if a step worked (e.g., output contains "hello").
- **Plugin**: Add new ways to run tasks (e.g., Kubernetes, custom).
- **Hook**: Run your own code when a step starts, succeeds, or fails.

.. image:: _static/iapetus-concept-diagram.png
   :alt: iapetus concepts diagram

**See also:** :doc:`api`, :doc:`yaml` 