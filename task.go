package iapetus

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
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
	Expected      Output              // Expected command output and behavior
	Actual        Output              // Actual command output and results
	SkipJsonNodes []string            // JSON nodes to ignore in comparisons
	Asserts       []func(*Task) error // Custom validation functions
	LogLevel      int
	logger        Logger // Add this field
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
func NewTask(name string, timeout time.Duration, level *LogLevel) *Task {
	if name == "" {
		name = "task-" + uuid.New().String()
	}
	return &Task{
		Name:    name,
		Timeout: timeout,
		logger:  NewDefaultLogger(level), // Initialize with default logger
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
		logLevel := LogLevel(t.LogLevel)
		t.SetLogger(&logLevel)
	}
	if t.Timeout == 0*time.Second {
		t.Timeout = 100 * time.Second
	}
	if t.Retries == 0 {
		t.Retries = 1
	}
	t.logger.Info("Running task: %s", t.Name)

	if t.PreRun != nil {
		t.logger.Debug("Starting pre-run hook for workflow: %s", t.Name)
		if err := t.PreRun(t); err != nil {
			t.logger.Error("Pre-run hook failed for workflow %s: %v", t.Name, err)
			return fmt.Errorf("pre-run hook failed for workflow %s: %v", t.Name, err)
		}
	}
	for attempt := 1; attempt <= t.Retries; attempt++ {
		t.logger.Debug("Attempt %d of %d for task: %s", attempt, t.Retries, t.Name)

		cmd := exec.Command(t.Command, t.Args...)
		cmd.Env = append(os.Environ(), t.Env...)

		// Create a context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), t.Timeout)
		defer cancel()
		cmd.WaitDelay = t.Timeout

		// Set the command to use our context
		cmd = exec.CommandContext(ctx, t.Command, t.Args...)
		cmd.Env = append(os.Environ(), t.Env...)

		t.logger.Debug("Command: %s", t.Command+" "+strings.Join(t.Args, " "))
		output, err := cmd.CombinedOutput()
		t.Actual.ExitCode = getExitCode(err)
		t.Actual.Output = string(output)

		if err != nil {
			t.Actual.Error = err.Error()
			if ctx.Err() == context.DeadlineExceeded {
				t.logger.Error("Task %s timed out after %v", t.Name, t.Timeout)
				return fmt.Errorf("task %s timed out after %v", t.Name, t.Timeout)
			}
			t.logger.Error("Error executing task %s: %v", t.Name, err)
		}

		for _, assert := range t.Asserts {
			if err := assert(t); err != nil {
				t.logger.Error("Assertion failed for task %s: %v", t.Name, err)
				if attempt < t.Retries {
					t.logger.Debug("Retrying task %s after failure", t.Name)
					time.Sleep(1 * time.Second)
					continue
				}
				fmt.Println("Command ", t.Command, strings.Join(t.Args, " "))
				return fmt.Errorf("assertion failed for task %s: %w", t.Name, err)
			}
		}
		break
	}

	if t.PostRun != nil {
		t.logger.Debug("Starting post-run hook for workflow: %s", t.Name)
		if err := t.PostRun(t); err != nil {
			t.logger.Error("Post-run hook failed for workflow %s: %v", t.Name, err)
			return fmt.Errorf("post-run hook failed for workflow %s: %v", t.Name, err)
		}
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

// Add method to set custom logger
func (t *Task) SetLogger(level *LogLevel) *Task {
	t.logger = NewDefaultLogger(level)
	return t
}

func (t *Task) SetRetries(retry int) *Task {
	t.Retries = retry
	return t
}
