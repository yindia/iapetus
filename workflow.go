package iapetus

import (
	"fmt"

	"github.com/google/uuid"
)

// Package iapetus provides workflow orchestration capabilities

// WorkflowError represents an error that occurred during workflow execution.
// It contains context about which step failed and in which workflow.
type WorkflowError struct {
	// StepName is the name of the task that failed
	StepName string
	// WorkflowName is the identifier of the workflow where the error occurred
	WorkflowName string
	// Err is the underlying error that caused the failure
	Err error
}

// Error implements the error interface for WorkflowError.
func (e *WorkflowError) Error() string {
	return fmt.Sprintf("error in step '%s' of workflow '%s': %v", e.StepName, e.WorkflowName, e.Err)
}

// Workflow represents a sequence of tasks to be executed in order.
// It provides hooks for pre and post-execution logic and maintains
// an ordered list of tasks to be executed sequentially.
type Workflow struct {
	// Name identifies the workflow. If empty, a UUID will be generated at runtime
	Name string // Name identifies the workflow
	// PreRun is executed before any tasks. It can be used for workflow initialization
	PreRun func(w *Workflow) error // PreRun is executed before any tasks
	// PostRun is executed after all tasks complete successfully
	PostRun func(w *Workflow) error // PostRun is executed after all tasks
	// Steps contains the ordered list of tasks to execute
	Steps []Task // Steps contains the ordered list of tasks to execute

	LogLevel int
	// logger handles workflow execution logging
	logger Logger
}

// NewWorkflow creates a new Workflow instance with the given name.
// If name is empty, a UUID-based name will be generated during execution.
func NewWorkflow(name string, level *LogLevel) *Workflow {
	return &Workflow{
		Name:   name,
		logger: NewDefaultLogger(level),
	}
}

// SetLogger allows users to configure a custom logger
func (w *Workflow) SetLogger(level *LogLevel) *Workflow {
	w.logger = NewDefaultLogger(level)
	return w
}

// Run executes the workflow by running all tasks in sequence.
// It handles pre-run and post-run hooks if defined.
// Returns an error if any step fails.
func (w *Workflow) Run() error {

	if w.logger == nil {
		logLevel := LogLevel(w.LogLevel)
		w.SetLogger(&logLevel)
	}
	w.logger.Info("Starting workflow: %s", w.Name)

	if w.Name == "" {
		w.Name = "workflow-" + uuid.New().String()
		w.logger.Debug("Generated new workflow name: %s", w.Name)
	}

	if w.PreRun != nil {
		w.logger.Debug("Starting pre-run hook for workflow: %s", w.Name)
		if err := w.PreRun(w); err != nil {
			w.logger.Error("Pre-run hook failed for workflow %s: %v", w.Name, err)
			return fmt.Errorf("pre-run hook failed for workflow %s: %v", w.Name, err)
		}
	}
	for _, task := range w.Steps {
		if err := task.Run(); err != nil {
			// Wrap the error with additional context
			wfErr := &WorkflowError{
				StepName:     task.Name,
				WorkflowName: w.Name,
				Err:          err,
			}
			w.logger.Error("Error: %v", wfErr)
			return wfErr
		}
	}

	if w.PostRun != nil {
		w.logger.Debug("Starting post-run hook for workflow: %s", w.Name)
		if err := w.PostRun(w); err != nil {
			w.logger.Error("Post-run hook failed for workflow %s: %v", w.Name, err)
			return fmt.Errorf("post-run hook failed for workflow %s: %v", w.Name, err)
		}
	}
	w.logger.Info("Completed workflow: %s", w.Name)
	return nil
}

// AddTask appends a new task to the workflow's sequence of steps.
// Tasks are executed in the order they are added.
// Returns the workflow to allow for method chaining.
func (w *Workflow) AddTask(task Task) *Workflow {
	w.Steps = append(w.Steps, task)
	return w
}
