# YAML Example: iapetus

This directory shows how to use **iapetus** with a YAML workflow file. You don't need to know Go to get started!

## What does this example do?
- Runs a simple workflow defined in `workflow_docker.yaml`
- Each step runs a command (like `echo hello`) and checks the result
- Shows how to use Docker as a backend (runs commands in containers)

## How to run the example

1. **Install Go** ([instructions](https://golang.org/dl/))
2. **Clone the repo**
   ```sh
   git clone https://github.com/yindia/iapetus.git
   cd iapetus/example/yaml
   ```
3. **Run the example**
   ```sh
   go run main_docker.go
   ```

You should see output for each step, showing if it passed or failed.

## How to modify the workflow
- Edit `workflow_docker.yaml` to add, remove, or change steps
- You can change commands, arguments, environment variables, and assertions
- Save the file and re-run `go run main_docker.go` to see your changes

## Troubleshooting
- If you see "command not found", make sure the command exists in the Docker image
- If you see Go errors, check that Go is installed and your `GOPATH` is set up
- For Docker errors, make sure Docker is running on your system

## More resources
- [iapetus README](../../README.md)
- [Go by Example](https://gobyexample.com/)
- [YAML Tutorial](https://learnxinyminutes.com/docs/yaml/) 