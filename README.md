# iapetus

A Go package for executing and validating command-line workflows with built-in assertions and error handling.

## Overview

The `iapetus` package provides:
- Structured workflow execution with sequential steps
- Built-in and custom assertions for validating outputs
- Retry mechanisms for flaky operations
- Fluent builder pattern for readable workflow construction


## Installation

```go
go get github.com/yindia/iapetus
```

## Usage

### Quick Start

```go
// Create a simple task
task := iapetus.NewTask("verify-service", 5*time.Second, 0).
    AddCommand("curl").
    AddArgs("-f", "http://localhost:8080").
    AddExpected(iapetus.Output{ExitCode: 0}).
    AddAssertion(iapetus.AssertByExitCode)

// Run the task
if err := task.Run(); err != nil {
    log.Fatalf("Task failed: %v", err)
}
```

### Defining a Step

A `Step` is defined with the command to run, its arguments, environment variables, and expected output. You can also add custom assertions to a step.

```go
step := iapetus.Step{
    Command: "kubectl",
    Args:    []string{"get", "pods", "-n", "default"},
    Env:     []string{},
    Expected: iapetus.Output{
        ExitCode: 1,
    },
    LogLevel: 0,
    Asserts: []func(*iapetus.Step) error{
        iapetus.AssertByExitCode,
        func(i *iapetus.Step) error {
            // TODO: Add custom assertion
            return nil
        },
    },
}

err := step.Run()
if err != nil {
    log.Fatalf("Failed to run step: %v", err)
}
```


### Define a Workflow

A `Workflow` consists of multiple steps that are executed in sequence. If any step fails its assertions, the workflow stops.

```go
workflow := iapetus.Workflow{
		Name: "Entire flow",
		PreRun: func(w *iapetus.Workflow) error {
			//# Do sonething
			return nil
		},
		LogLevel: 1,
		Steps: []iapetus.Task{
			{
				Name:    "kubectl-create-ns",
				Command: "kubectl",
				Args:    []string{"create", "ns", ns},
				Env:     []string{},
				Expected: iapetus.Output{
					ExitCode: 0,
				},
				Asserts: []func(*iapetus.Task) error{
					iapetus.AssertByExitCode,
				},
			},
			{
				Name:    "kubectl-create-deployment",
				Command: "kubectl",
				Args:    []string{"create", "deployment", "test", "--image", "nginx", "--replicas", "30", "-n", ns},
				Env:     []string{},
				Expected: iapetus.Output{
					ExitCode: 0,
				},
				Asserts: []func(*iapetus.Task) error{
					iapetus.AssertByExitCode,
				},
			},
			{
				Name:    "kubectl-get-pods-with-deployment",
				Command: "kubectl",
				Args:    []string{"get", "pods", "-n", ns, "-o", "json"},
				Env:     []string{},
				Retries: 1,
				Expected: iapetus.Output{
					ExitCode: 0,
				},
				Asserts: []func(*iapetus.Task) error{
					iapetus.AssertByExitCode,
					func(s *iapetus.Task) error {
						deployment := &appsv1.DeploymentList{}
						err := json.Unmarshal([]byte(s.Actual.Output), &deployment)
						if err != nil {
							return fmt.Errorf("failed to unmarshal deployment specs: %w", err)
						}
						if len(deployment.Items) == 1 {
							return fmt.Errorf("deployment length should be 1")
						}
						for _, item := range deployment.Items {
							if item.Name == "test" {
								for _, container := range item.Spec.Template.Spec.Containers {
									if container.Image != "nginx" {
										return fmt.Errorf("container image should be nginx")
									}
								}
								if item.Status.Replicas != *item.Spec.Replicas {
									return fmt.Errorf("deployment replicas do not match desired state")
								}
							}
						}
						return nil
					},
				},
			},
		},
}

err := workflow.Run()
if err != nil {
    log.Fatalf("Failed to run workflow: %v", err)
}
```

### Builder Pattern

The package provides a fluent builder pattern for creating tasks and workflows, making it easier to construct complex workflows with a more readable syntax:

```go
// Create a task with timeout and log level
step1 := iapetus.NewTask("create cluster", 10*time.Second, 0).
    AddCommand("kind").
    AddArgs("create", "cluster").
    AddExpected(iapetus.Output{
        ExitCode: 0,
    }).
    AddAssertion(iapetus.AssertByExitCode)

// Create another task
step2 := iapetus.NewTask("verify pods", 10*time.Second, 0).
    AddCommand("kubectl").
    AddArgs("get", "pods").
    AddExpected(iapetus.Output{
        ExitCode: 0,
    }).
    AddAssertion(iapetus.AssertByExitCode)

// Combine tasks into a workflow
workflow := iapetus.NewWorkflow("cluster-setup", 0).
    AddTask(step2).
    AddPreRun(func(w *Workflow) error {
        if err := step1.Run(); err != nil {
            return err
        }
    })

// Run the workflow
if err := workflow.Run(); err != nil {
    log.Fatalf("Workflow failed: %v", err)
}
```

### Assertions

The package provides several built-in assertion functions:

- `AssertByExitCode`: Validates the exit code of a step.
- `AssertByOutputString`: Compares the actual output string with the expected output.
- `AssertByOutputJson`: Compares JSON outputs, allowing for specific node skipping.
- `AssertByContains`: Checks if the actual output contains specific strings.
- `AssertByError`: Validates the error message.
- `AssertByRegexp`: Validates the output with regx


### Custom Assertions

You can add custom assertions to a step using the `AddAssertion` method. A custom assertion is a function that takes a `*Step` as an argument and returns an error if the assertion fails.

```go
step.AddAssertion(func(i *iapetus.Step) error {
    if i.Actual.ExitCode != 0 {
        return fmt.Errorf("expected exit code 0, but got %d", i.Actual.ExitCode)
    }
    return nil
})
```

### Key Features

#### Retries
Tasks can be configured with retries for handling transient failures:

```go
task := iapetus.NewTask("flaky-operation", timeout, 0).
    AddCommand("some-command").
    SetRetries(3)  // Will retry up to 3 times
```

#### Built-in Assertions
```go
task.AddAssertion(iapetus.AssertByExitCode).        // Check exit code
    AddAssertion(iapetus.AssertByContains("Ready")) // Check output contains
```

Available assertions:
- `AssertByExitCode`: Validates command exit code
- `AssertByOutputString`: Exact string matching
- `AssertByOutputJson`: JSON comparison with node skipping
- `AssertByContains`: Substring presence check
- `AssertByError`: Error message validation
- `AssertByRegexp`: Regular expression matching

#### Custom Assertions
```go
task.AddAssertion(func(t *iapetus.Task) error {
    if !strings.Contains(t.Actual.Output, "success") {
        return fmt.Errorf("expected 'success' in output")
    }
    return nil
})
```

## Contributing

Contributions to the `iapetus` package are welcome. Please submit issues or pull requests via the project's repository.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
