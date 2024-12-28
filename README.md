# iapetus

The `iapetus` package is designed to facilitate the execution and validation of command-line steps within a workflow. It provides a structured way to define, run, and assert the outcomes of various steps, making it ideal for testing and automation tasks.

## Overview

The package consists of three main components:

1. **Step**: Represents a single command-line step with its expected and actual outputs.
2. **Workflow**: Manages a sequence of steps and executes them in order.
3. **Assertions**: Provides various assertion functions to validate the outcomes of steps.

## Installation

To use the `iapetus` package, you need to import it into your Go project:

```go
go get github.com/yindia/iapetus
```

## Usage

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
    Asserts: []func(*iapetus.Step) error{
        iapetus.AssertByExitCode,
        func(i *iapetus.Step) error {
            // TODO: Add custom assertion
            return nil
        },
    },
}
```

### Running a Step

To execute a step and validate its output, use the `Run` and `Assert` methods:

```go
err := step.Run()
if err != nil {
    log.Fatalf("Failed to run step: %v", err)
}
```


### Using a Workflow

A `Workflow` consists of multiple steps that are executed in sequence. If any step fails its assertions, the workflow stops.

```go
workflow := iapetus.Workflow{
    Steps: []iapetus.Step{step1, step2, step3},
}

err := workflow.Run()
if err != nil {
    log.Fatalf("Failed to run workflow: %v", err)
}
```
### Assertions

The package provides several built-in assertion functions:

- `AssertByExitCode`: Validates the exit code of a step.
- `AssertByOutputString`: Compares the actual output string with the expected output.
- `AssertByOutputJson`: Compares JSON outputs, allowing for specific node skipping.
- `AssertByContains`: Checks if the actual output contains specific strings.
- `AssertByError`: Validates the error message.

You can add custom assertions to a step using the `AddAssertion` method.

```go
step.AddAssertion(iapetus.AssertByExitCode)
step.AddAssertion(iapetus.AssertByOutputString)
step.AddAssertion(iapetus.AssertByOutputJson)
step.AddAssertion(iapetus.AssertByContains)
step.AddAssertion(iapetus.AssertByError)
step.AddAssertion(func(i *iapetus.Step) error {
    // TODO: Add custom assertion
    return nil
})
```

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

## Contributing

Contributions to the `iapetus` package are welcome. Please submit issues or pull requests via the project's repository.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
