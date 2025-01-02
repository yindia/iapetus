package iapetus

import (
	"fmt"
	"log"

	"github.com/google/uuid"
)

// Define a custom error type for workflow errors
type WorkflowError struct {
	StepName     string
	WorkflowName string
	Err          error
}

func (e *WorkflowError) Error() string {
	return fmt.Sprintf("error in step '%s' of workflow '%s': %v", e.StepName, e.WorkflowName, e.Err)
}

type Workflow struct {
	Name    string
	PreRun  func(w *Workflow) error
	PostRun func(w *Workflow) error
	Steps   []Task
}

func NewWorkflow(name string) *Workflow {
	return &Workflow{Name: name}
}

func (w *Workflow) Run() error {
	log.Printf("Starting workflow: %s", w.Name)

	if w.Name == "" {
		w.Name = "workflow-" + uuid.New().String()
		log.Printf("Generated new workflow name: %s", w.Name)
	}

	if w.PreRun != nil {
		log.Println("Starting workflow Pre Run")
		if err := w.PreRun(w); err != nil {
			return fmt.Errorf("Failed pre run for the workflow %s got error %s "+w.Name, err.Error())
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
			log.Printf("Error: %v", wfErr)
			return wfErr
		}
	}

	if w.PostRun != nil {
		log.Println("Starting workflow Post Run")
		if err := w.PostRun(w); err != nil {
			return fmt.Errorf("Failed post run for the workflow %s got error %s "+w.Name, err.Error())
		}
	}
	log.Printf("Completed workflow: %s", w.Name)
	return nil
}

func (w *Workflow) AddTask(task Task) *Workflow {
	w.Steps = append(w.Steps, task)
	return w
}
