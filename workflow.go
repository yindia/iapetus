package iapetus

import (
	"log"

	"github.com/google/uuid"
)

type Workflow struct {
	Name  string
	Steps []Step
}

func NewWorkflow(name string) *Workflow {
	return &Workflow{Name: name}
}

func (w *Workflow) Run() error {
	if w.Name == "" {
		w.Name = "workflow-" + uuid.New().String()
		log.Printf("Generated new workflow name: %s", w.Name)
	}

	for _, step := range w.Steps {
		if err := step.Run(); err != nil {
			log.Printf("Running step: %s in workflow: %s got error: %v", step.Name, w.Name, err)
			return err
		}
	}
	return nil
}

func (w *Workflow) AddStep(step Step) *Workflow {
	w.Steps = append(w.Steps, step)
	return w
}