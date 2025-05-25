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

// Task represents a configurable command execution unit that can be run with retries
// and validated against expected outputs and custom assertiont.
type Task struct {
	Name          string              // Unique identifier for the task
	Command       string              // The command to execute
	Retries       int                 // Number of retry attempts if assertions fail
	Args          []string            // Command line arguments
	Timeout       time.Duration       // Maximum execution time
	Env           []string            // Additional environment variables
	WorkingDir    string              // Working dir
	Expected      Output              // Expected command output and behavior
	Depends       []string            // Dependencies for the task
	Actual        Output              // Actual command output and results
	SkipJsonNodes []string            // JSON nodes to ignore in comparisons
	Asserts       []func(*Task) error // Custom validation functions
	logger        *zap.Logger
	// PreRun is executed before any tasks. It can be used for task initialization
	PreRun func(w *Task) error // PreRun is executed before any tasks
	// PostRun is executed after all tasks complete successfully
	PostRun func(w *Task) error // PostRun is executed after all tasks
}

// Output holds the execution results of a command, including its exit code,
// standard output, error output, and expected content.
type Output struct {
	ExitCode int      // Process exit code
	Output   string   // Combined stdout and stderr
	Error    string   // Error message if execution failed
	Contains []string // Strings that should be present in the output
	Patterns []string // Regular expression pattern to match against the output
}

// NewTask creates a new Task instance with the specified name and timeout.
// If name is empty, a UUID-based name will be generated.
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
	}
}

// Run executes the task with configured retries and assertiont.
// It captures the command output and runs all registered assertiont.
// Returns an error if any assertion fails after all retry attemptt.
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
// Assertions are run in the order they are added after the command completet.
func (t *Task) AddAssertion(assert func(*Task) error) *Task {
	t.Asserts = append(t.Asserts, assert)
	return t
}

func (t *Task) AddContains(contains ...string) *Task {
	t.Expected.Contains = append(t.Expected.Contains, contains...)
	return t
}

func (t *Task) AddEnv(env ...string) *Task {
	t.Env = append(t.Env, env...)
	return t
}

func (t *Task) AddArgs(args ...string) *Task {
	t.Args = append(t.Args, args...)
	return t
}

func (t *Task) AddSkipJsonNodes(skipJsonNodes ...string) *Task {
	t.SkipJsonNodes = append(t.SkipJsonNodes, skipJsonNodes...)
	return t
}

func (t *Task) AddExpected(expected Output) *Task {
	t.Expected = expected
	return t
}

func (t *Task) AddCommand(command string) *Task {
	t.Command = command
	return t
}

func (t *Task) SetRetries(retry int) *Task {
	t.Retries = retry
	return t
}
