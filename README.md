# iapetus 🚀

**The Open Source Workflow Engine for DevOps, CI/CD, and Automation**

[![Go Reference](https://pkg.go.dev/badge/github.com/yindia/iapetus.svg)](https://pkg.go.dev/github.com/yindia/iapetus)
[![Go Report Card](https://goreportcard.com/badge/github.com/yindia/iapetus)](https://goreportcard.com/report/github.com/yindia/iapetus)
[![codecov](https://codecov.io/gh/yindia/iapetus/graph/badge.svg?token=6S99FUSPOC)](https://codecov.io/gh/yindia/iapetus)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

---
> **Automate, test, and orchestrate anything that runs in a shell, container, or cloud.**  
> **No YAML hell. No vendor lock-in. 100% open source.**

---

# ⚠️ API subject to change! iapetus is in heavy development. Expect breaking changes. ⚠️


## ✨ Why iapetus?

- ⚡️ **Lightning-fast**: Parallel, dependency-aware execution
- 🔌 **Pluggable**: Bash, Docker, and custom backends
- 🧪 **Assertions**: Output, exit code, JSON, regex, and more
- 📝 **YAML or Go**: Use as config or code
- 🛡️ **Battle-tested**: For CI/CD, DevOps, and E2E testing

---

## 🚀 Demo

![demo](https://github.com/user-attachments/assets/521ce88d-609d-44bb-a605-244eb80429f9)

---

## ⚡️ Quickstart

```sh
git clone https://github.com/yindia/iapetus.git
cd iapetus/example/yaml
go run main_docker.go
```

---

## 📝 Example: YAML Workflow

```yaml
name: hello-world
steps:
  - name: say-hello
    command: echo
    args: ["Hello, iapetus!"]
    raw_asserts:
      - output_contains: iapetus
```

---

## 💻 Example: Go API

```go
task := iapetus.NewTask("say-hello", 2*time.Second, nil).
    AddCommand("echo").
    AddArgs("Hello, iapetus!").
    AssertOutputContains("iapetus")
workflow := iapetus.NewWorkflow("hello-world", zap.NewNop()).AddTask(*task)
workflow.Run()
```

---

## 🧩 Features

- 🔄 **Parallel, dependency-aware execution**
- ✅ **Built-in & custom assertions**
- ⏱️ **Retries, timeouts, env vars, container images**
- 🔌 **Plugin backends**: Bash, Docker, and more
- 🪝 **Hooks for logging, metrics, and custom logic**
- 📊 **Beautiful logs and error reporting**

---

## 🤝 Contributing

We welcome PRs, issues, and feedback! See [Contributing Guide](https://iapetus.readthedocs.io/en/latest/contributing.html).

---

## 📜 License

MIT

---

🌟 **Star iapetus if you love it!**
