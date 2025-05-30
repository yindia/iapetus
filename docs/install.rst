Installation Guide
==================

Prerequisites
-------------

- **Go** (version 1.18 or newer recommended)
- **(Optional) Docker** (for running tasks in containers)
- **(Optional) [pixi](https://pixi.sh/)** for reproducible development environments

Step 1: Install Go
------------------

Download and install Go from https://golang.org/dl/

.. code-block:: shell

   go version

Step 2: Clone the Repository
----------------------------

.. code-block:: shell

   git clone https://github.com/yindia/iapetus.git
   cd iapetus

Step 3: Install Dependencies
----------------------------

.. code-block:: shell

   go mod tidy

Step 4: (Optional) Use pixi for Reproducible Environments
---------------------------------------------------------

.. code-block:: shell

   pixi shell

Step 5: Run Tests to Verify Installation
----------------------------------------

.. code-block:: shell

   go test ./...

Step 6: (Optional) Run Linter
-----------------------------

.. code-block:: shell

   golangci-lint run ./...

Troubleshooting
---------------

- **Go not found**: Ensure Go is installed and in your PATH
- **Docker errors**: Make sure Docker is installed and running
- **Dependency errors**: Run `go mod tidy` to resolve
- **Permission denied**: Check file and Docker permissions

For more help, see the FAQ in the main README or open an issue on GitHub. 