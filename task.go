package iapetus

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Task represents a configurable command execution unit that can be run with retries,
// validated against expected outputs, and extended with custom assertions and hooks.
type Task struct {
	// Name is a unique identifier for the task.
	Name string // Unique identifier for the task
	// Command is the command to execute.
	Command string // The command to execute
	// Retries is the number of retry attempts if assertions fail.
	Retries int // Number of retry attempts if assertions fail
	// Args are command line arguments for the command.
	Args []string // Command line arguments
	// Timeout is the maximum execution time for the task.
	Timeout time.Duration // Maximum execution time
	// Env is a list of additional environment variables (KEY=VALUE).
	Env []string `json:"env" yaml:"env"` // Additional environment variables (KEY=VALUE)
	// EnvMap is an alternative environment variable representation (key-value map).
	EnvMap map[string]string `json:"env_map" yaml:"env_map"` // Alternative env representation (key-value)
	// Image is the container image to use for the task (optional).
	Image string `json:"image" yaml:"image"` // Container image to use for the task (optional)
	// WorkingDir is the working directory for the command.
	WorkingDir string // Working dir
	// Depends lists the names of tasks this task depends on.
	Depends []string // Dependencies for the task
	// Actual holds the actual output and results of the command execution.
	Actual Output // Actual command output and results
	// Asserts is a list of custom validation functions (assertions).
	Asserts []func(*Task) error // Custom validation functions
	// logger is the zap logger used for this task.
	logger *zap.Logger
	// PreRun is executed before the task runs. Can be used for task initialization
	PreRun func(w *Task) error // PreRun is executed before any tasks
	// PostRun is executed after the task completes successfully.
	PostRun func(w *Task) error // PostRun is executed after all tasks complete successfully
}

// Output holds the execution results of a command, including its exit code,
// standard output, error output, and expected content.
type Output struct {
	// ExitCode is the process exit code.
	ExitCode int // Process exit code
	// Output is the combined stdout and stderr.
	Output string // Combined stdout and stderr
	// Error is the error message if execution failed.
	Error string // Error message if execution failed
	// Contains lists strings that should be present in the output.
	Contains []string // Strings that should be present in the output
	// Patterns are regular expression patterns to match against the output.
	Patterns []string // Regular expression pattern to match against the output
}

// NewTask creates a new Task instance with the specified name and timeout.
// If name is empty, a UUID-based name will be generated. If logger is nil, a production zap logger is used.
func NewTask(name string, timeout time.Duration, logger *zap.Logger) *Task {
	if name == "" {
		name = "task-" + uuid.New().String()
	}
	if logger == nil {
		logger, _ = zap.NewProduction()
	}
	return &Task{
		Name:    name,
		Timeout: timeout,
		logger:  logger,
		EnvMap:  make(map[string]string), // Initialize EnvMap
	}
}

// Run executes the task with configured retries and assertions.
// It captures the command output and runs all registered assertions.
// Returns an error if any assertion fails after all retry attempts.
func (t *Task) Run() error {
	if t.Name == "" {
		t.Name = "task-" + uuid.New().String()
	}
	if t.logger == nil {
		t.logger, _ = zap.NewProduction()
	}
	if t.Timeout == 0*time.Second {
		t.Timeout = 100 * time.Second
	}
	if t.Retries == 0 {
		t.Retries = 1
	}
	t.logger.Info("Running task", zap.String("task", t.Name))
	if t.PreRun != nil {
		t.logger.Debug("Starting pre-run hook", zap.String("task", t.Name))
		if err := t.PreRun(t); err != nil {
			t.logger.Error("Pre-run hook failed", zap.String("task", t.Name), zap.Error(err))
			return fmt.Errorf("pre-run hook failed for task %s: %v", t.Name, err)
		}
	}
	var lastErr error
	for attempt := 1; attempt <= t.Retries; attempt++ {
		t.logger.Debug("Attempt", zap.Int("attempt", attempt), zap.Int("retries", t.Retries), zap.String("task", t.Name))
		if err := t.executeCommand(); err != nil {
			lastErr = err
			if attempt < t.Retries {
				t.logger.Debug("Retrying task after failure", zap.String("task", t.Name))
				time.Sleep(1 * time.Second)
				continue
			}
			return fmt.Errorf("task %s failed after %d attempts: %w", t.Name, t.Retries, err)
		}
		return nil
	}
	if t.PostRun != nil {
		t.logger.Debug("Starting post-run hook", zap.String("task", t.Name))
		if err := t.PostRun(t); err != nil {
			t.logger.Error("Post-run hook failed", zap.String("task", t.Name), zap.Error(err))
			return fmt.Errorf("post-run hook failed for task %s: %v", t.Name, err)
		}
	}
	return lastErr
}

// executeCommand handles a single execution attempt
func (t *Task) executeCommand() error {
	ctx, cancel := context.WithTimeout(context.Background(), t.Timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, t.Command, t.Args...)
	cmd.Env = append(os.Environ(), t.Env...)
	if t.WorkingDir != "" {
		cmd.Dir = t.WorkingDir
	}
	t.logger.Debug("Command", zap.String("cmd", t.Command+" "+strings.Join(t.Args, " ")))
	output, err := cmd.CombinedOutput()
	t.Actual.Output = string(output)
	t.Actual.ExitCode = getExitCode(err)
	if err != nil {
		t.Actual.Error = err.Error()
		if ctx.Err() == context.DeadlineExceeded {
			t.logger.Error("Task timed out", zap.String("task", t.Name), zap.Duration("timeout", t.Timeout))
			return fmt.Errorf("task %s timed out after %v", t.Name, t.Timeout)
		}
		t.logger.Error("Error executing task", zap.String("task", t.Name), zap.Error(err))
	}
	// Use RunAssertions to aggregate all assertion errors
	err = RunAssertions(t)
	if err != nil {
		t.logger.Error("Assertion(s) failed", zap.String("task", t.Name), zap.Error(err))
		return err
	}
	return nil
}

// AddAssertion registers a new assertion function to validate the task execution.
// Assertions are run in the order they are added after the command completes.
func (t *Task) AddAssertion(assert func(*Task) error) *Task {
	t.Asserts = append(t.Asserts, assert)
	return t
}

// AddEnv appends environment variables to the task.
func (t *Task) AddEnv(env ...string) *Task {
	t.Env = append(t.Env, env...)
	return t
}

// AddArgs appends command line arguments to the task.
func (t *Task) AddArgs(args ...string) *Task {
	t.Args = append(t.Args, args...)
	return t
}

// AddCommand sets the command to execute for the task.
func (t *Task) AddCommand(command string) *Task {
	t.Command = command
	return t
}

// SetRetries sets the number of retry attempts for the task.
func (t *Task) SetRetries(retry int) *Task {
	t.Retries = retry
	return t
}

// AssertExitCode adds an assertion that checks the exit code of the task.
func (t *Task) AssertExitCode(code int) *Task {
	return t.AddAssertion(AssertExitCode(code))
}

// AssertOutputContains adds an assertion that checks if output contains a substring.
func (t *Task) AssertOutputContains(substr string) *Task {
	return t.AddAssertion(AssertOutputContains(substr))
}

// AssertOutputEquals adds an assertion that checks if output matches exactly.
func (t *Task) AssertOutputEquals(expected string) *Task {
	return t.AddAssertion(AssertOutputEquals(expected))
}

// AssertOutputJsonEquals adds an assertion that checks if output JSON matches expected JSON.
func (t *Task) AssertOutputJsonEquals(expected string, skipJsonNodes ...string) *Task {
	return t.AddAssertion(AssertOutputJsonEquals(expected, skipJsonNodes...))
}

// AssertOutputMatchesRegexp adds an assertion that checks if output matches a regexp.
func (t *Task) AssertOutputMatchesRegexp(pattern string) *Task {
	return t.AddAssertion(AssertOutputMatchesRegexp(pattern))
}

// Expect returns a new TaskAssertionBuilder for chaining assertions in a fluent style.
func (t *Task) Expect() *TaskAssertionBuilder {
	return &TaskAssertionBuilder{task: t}
}

// TaskAssertionBuilder allows chaining assertion methods in a fluent style.
// Usage: task.Expect().ExitCode(0).OutputContains("foo").Done()
type TaskAssertionBuilder struct {
	task *Task
}

// ExitCode adds an exit code assertion to the builder.
func (b *TaskAssertionBuilder) ExitCode(code int) *TaskAssertionBuilder {
	b.task.AssertExitCode(code)
	return b
}

// OutputContains adds an output substring assertion to the builder.
func (b *TaskAssertionBuilder) OutputContains(substr string) *TaskAssertionBuilder {
	b.task.AssertOutputContains(substr)
	return b
}

// OutputEquals adds an output equality assertion to the builder.
func (b *TaskAssertionBuilder) OutputEquals(expected string) *TaskAssertionBuilder {
	b.task.AssertOutputEquals(expected)
	return b
}

// OutputJsonEquals adds a JSON output equality assertion to the builder.
func (b *TaskAssertionBuilder) OutputJsonEquals(expected string, skipJsonNodes ...string) *TaskAssertionBuilder {
	b.task.AssertOutputJsonEquals(expected, skipJsonNodes...)
	return b
}

// OutputMatchesRegexp adds a regexp output assertion to the builder.
func (b *TaskAssertionBuilder) OutputMatchesRegexp(pattern string) *TaskAssertionBuilder {
	b.task.AssertOutputMatchesRegexp(pattern)
	return b
}

// Done returns the parent Task for further chaining.
func (b *TaskAssertionBuilder) Done() *Task {
	return b.task
}

// AddImage sets the container image for the task.
func (t *Task) AddImage(image string) *Task {
	t.Image = image
	return t
}

// AddEnvMap sets the EnvMap for the task (overwrites existing).
func (t *Task) AddEnvMap(envMap map[string]string) *Task {
	t.EnvMap = envMap
	return t
}
