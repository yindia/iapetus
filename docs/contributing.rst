Contributing Guide
==================

Thank you for your interest in contributing to iapetus! 🎉

We welcome all contributions—code, documentation, tests, bug reports, feature requests, and ideas.

Development Environment Setup
----------------------------

1. **Clone the repository**

   .. code-block:: shell

      git clone https://github.com/yindia/iapetus.git
      cd iapetus

2. **Install Go** ([instructions](https://golang.org/dl/))

3. **(Optional) Use [pixi](https://pixi.sh/)**

   .. code-block:: shell

      pixi shell

4. **Install dependencies**

   .. code-block:: shell

      go mod tidy

Testing & Linting
-----------------

- **Run all tests:**

  .. code-block:: shell

     go test ./...

- **Run linter:**

  .. code-block:: shell

     golangci-lint run ./...

- **(With pixi):**

  .. code-block:: shell

     pixi run go test ./...
     pixi run golangci-lint run ./...

Code Style & Guidelines
-----------------------

- Follow idiomatic Go style (`gofmt`, `golangci-lint`)
- Write clear, descriptive commit messages (conventional commits preferred)
- Add/maintain unit tests for all features and bugfixes
- Document public APIs with GoDoc comments
- Keep PRs focused and small when possible
- Be kind and constructive in code reviews and discussions

Submitting Issues & Feature Requests
------------------------------------

- **Bugs:** Please include steps to reproduce, expected vs actual behavior, and environment info.
- **Features:** Describe the problem, your proposed solution, and alternatives considered.
- Use the GitHub issue templates when possible.

Submitting Pull Requests
------------------------

1. Fork the repo and create your branch from `main`.
2. Add your changes and tests.
3. Run tests and linter locally.
4. Open a pull request with a clear description of your changes.
5. Reference any related issues in your PR description.

Documentation
-------------

- Update or add documentation for any new features or changes.
- Keep the README, USAGE.md, and REFERENCE.md up to date as needed.
- Add code comments for exported functions, types, and complex logic.

Community & Conduct
-------------------

- Be respectful and welcoming to all contributors.
- Follow the Contributor Covenant Code of Conduct.
- Join GitHub Discussions for questions, ideas, or to connect with the community.

Thank you for helping make iapetus better! 🚀 