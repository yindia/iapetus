package main

import (
	"fmt"
	"log"

	"github.com/yindia/iapetus"
)

func main() {
	wf, err := iapetus.LoadWorkflowFromYAML("workflow_docker.yaml")
	if err != nil {
		log.Fatalf("Failed to load workflow from YAML: %v", err)
	}
	fmt.Println("Loaded workflow:", wf.Name)
	if err := wf.Run(); err != nil {
		log.Fatalf("Workflow failed: %v", err)
	}
	fmt.Println("Workflow completed successfully!")
}
