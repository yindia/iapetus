package iapetus

import (
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
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

	Image  string            `json:"image" yaml:"image"`     // Container image for the workflow (optional)
	EnvMap map[string]string `json:"env_map" yaml:"env_map"` // Environment variables for the workflow (key-value)

	logger *zap.Logger

	// Observability hooks (can be set for testing or custom behavior)
	OnTaskStartHooks    []func(*Task)
	OnTaskSuccessHooks  []func(*Task)
	OnTaskFailureHooks  []func(*Task, error)
	OnTaskCompleteHooks []func(*Task)
}

// NewWorkflow creates a new Workflow instance with the given name.
// If name is empty, a UUID-based name will be generated during execution.
func NewWorkflow(name string, logger *zap.Logger) *Workflow {
	if logger == nil {
		logger, _ = zap.NewProduction()
	}
	return &Workflow{
		Name:                name,
		logger:              logger,
		EnvMap:              make(map[string]string), // Initialize EnvMap
		OnTaskStartHooks:    []func(*Task){},
		OnTaskSuccessHooks:  []func(*Task){},
		OnTaskFailureHooks:  []func(*Task, error){},
		OnTaskCompleteHooks: []func(*Task){},
	}
}

// Run executes the workflow by running all tasks in sequence.
// It handles pre-run and post-run hooks if defined.
// Returns an error if any step fails.
func (w *Workflow) Run() error {
	w.logger.Info("Starting workflow", zap.String("workflow", w.Name))
	if w.Name == "" {
		w.Name = "workflow-" + uuid.New().String()
		w.logger.Debug("Generated new workflow name", zap.String("workflow", w.Name))
	}
	if w.PreRun != nil {
		w.logger.Debug("Starting pre-run hook", zap.String("workflow", w.Name))
		if err := w.PreRun(w); err != nil {
			w.logger.Error("Pre-run hook failed", zap.String("workflow", w.Name), zap.Error(err))
			return fmt.Errorf("pre-run hook failed for workflow %s: %v", w.Name, err)
		}
	}
	dag := NewDag()
	for i := range w.Steps {
		task := &w.Steps[i]
		if err := dag.AddTask(task); err != nil {
			w.logger.Error("Failed to add task to DAG", zap.String("task", task.Name), zap.Error(err))
			return &WorkflowError{
				StepName:     task.Name,
				WorkflowName: w.Name,
				Err:          err,
			}
		}
	}
	if err := dag.Validate(); err != nil {
		w.logger.Error("DAG validation failed", zap.Error(err))
		return &WorkflowError{
			StepName:     "DAG",
			WorkflowName: w.Name,
			Err:          err,
		}
	}
	err := w.runParallelDAG(dag)
	if w.PostRun != nil {
		w.logger.Debug("Starting post-run hook", zap.String("workflow", w.Name))
		if postErr := w.PostRun(w); postErr != nil {
			w.logger.Error("Post-run hook failed", zap.String("workflow", w.Name), zap.Error(postErr))
			return fmt.Errorf("post-run hook failed for workflow %s: %v", w.Name, postErr)
		}
	}
	w.logger.Info("Completed workflow", zap.String("workflow", w.Name))
	return err
}

// runParallelDAG executes the tasks in the DAG in parallel according to dependencies.
// Returns the first error encountered, or nil if all tasks succeed.
func (w *Workflow) runParallelDAG(dag *DAG) error {
	order, err := dag.GetTopologicalOrder()
	if err != nil {
		w.logger.Error("DAG topological sort failed", zap.Error(err))
		return &WorkflowError{
			StepName:     "DAG",
			WorkflowName: w.Name,
			Err:          err,
		}
	}
	scheduler := newDagScheduler(w, order)
	return scheduler.run()
}

// Add hook registration methods
func (w *Workflow) AddOnTaskStartHook(hook func(*Task)) *Workflow {
	w.OnTaskStartHooks = append(w.OnTaskStartHooks, hook)
	return w
}
func (w *Workflow) AddOnTaskSuccessHook(hook func(*Task)) *Workflow {
	w.OnTaskSuccessHooks = append(w.OnTaskSuccessHooks, hook)
	return w
}
func (w *Workflow) AddOnTaskFailureHook(hook func(*Task, error)) *Workflow {
	w.OnTaskFailureHooks = append(w.OnTaskFailureHooks, hook)
	return w
}
func (w *Workflow) AddOnTaskCompleteHook(hook func(*Task)) *Workflow {
	w.OnTaskCompleteHooks = append(w.OnTaskCompleteHooks, hook)
	return w
}

// Observability hooks (call all registered hooks)
func (w *Workflow) OnTaskStart(task *Task) {
	for _, hook := range w.OnTaskStartHooks {
		hook(task)
	}
}
func (w *Workflow) OnTaskSuccess(task *Task) {
	for _, hook := range w.OnTaskSuccessHooks {
		hook(task)
	}
}
func (w *Workflow) OnTaskFailure(task *Task, err error) {
	for _, hook := range w.OnTaskFailureHooks {
		hook(task, err)
	}
}
func (w *Workflow) OnTaskComplete(task *Task) {
	for _, hook := range w.OnTaskCompleteHooks {
		hook(task)
	}
}

// AddTask appends a new task to the workflow's sequence of steps.
// Tasks are executed in the order they are added.
// Returns the workflow to allow for method chaining.
func (w *Workflow) AddTask(task Task) *Workflow {
	w.Steps = append(w.Steps, task)
	return w
}

func (w *Workflow) AddPreRun(p func(w *Workflow) error) *Workflow {
	w.PreRun = p
	return w
}

func (w *Workflow) AddPostRun(p func(w *Workflow) error) *Workflow {
	w.PostRun = p
	return w
}

// AddImage sets the container image for the workflow
func (w *Workflow) AddImage(image string) *Workflow {
	w.Image = image
	return w
}

// AddEnvMap sets the EnvMap for the workflow (overwrites existing)
func (w *Workflow) AddEnvMap(envMap map[string]string) *Workflow {
	w.EnvMap = envMap
	return w
}
