package main

import (
	"fmt"
	"os"
	"time"

	"github.com/yindia/iapetus"
	"go.uber.org/zap"
)

func main() {
	image := os.Getenv("DOCKER_IMAGE")
	if image == "" {
		image = "alpine"
	}
	busybox := "busybox"

	workflow := iapetus.NewWorkflow("Docker CLI Complex Example", zap.NewNop())

	// Register hooks for observability
	workflow.AddOnTaskStartHook(func(task *iapetus.Task) {
		fmt.Printf("[HOOK] Task started: %s\n", task.Name)
	})
	workflow.AddOnTaskSuccessHook(func(task *iapetus.Task) {
		fmt.Printf("[HOOK] Task succeeded: %s\n", task.Name)
	})
	workflow.AddOnTaskFailureHook(func(task *iapetus.Task, err error) {
		fmt.Printf("[HOOK] Task failed: %s, error: %v\n", task.Name, err)
	})
	workflow.AddOnTaskCompleteHook(func(task *iapetus.Task) {
		fmt.Printf("[HOOK] Task completed: %s\n", task.Name)
	})

	// Pull images
	pullAlpine := &iapetus.Task{
		Name:    "Pull Alpine Image",
		Command: "docker",
		Args:    []string{"pull", image},
		Asserts: []func(*iapetus.Task) error{
			iapetus.AssertExitCode(0),
		},
	}
	pullBusybox := &iapetus.Task{
		Name:    "Pull Busybox Image",
		Command: "docker",
		Args:    []string{"pull", busybox},
		Asserts: []func(*iapetus.Task) error{
			iapetus.AssertExitCode(0),
		},
	}

	// Run echo container (depends on both images)
	runEcho := iapetus.NewTask("Run Echo Container", 10*time.Second, nil).
		AddCommand("docker").
		AddArgs("run", "--rm", "--name", "hello-docker", image, "echo", "Hello from Docker!").
		AssertExitCode(0).
		AssertOutputContains("Hello from Docker!")
	runEcho.Depends = []string{"Pull Alpine Image", "Pull Busybox Image"}

	// Run sleep container (depends only on busybox)
	runSleep := iapetus.NewTask("Run Sleep Container", 10*time.Second, nil).
		AddCommand("docker").
		AddArgs("run", "-d", "--name", "sleep-docker", busybox, "sleep", "5").
		AssertExitCode(0)
	runSleep.Depends = []string{"Pull Busybox Image"}

	// Remove echo container
	removeEcho := &iapetus.Task{
		Name:    "Remove Echo Container",
		Command: "docker",
		Args:    []string{"rm", "-f", "hello-docker"},
		Asserts: []func(*iapetus.Task) error{
			iapetus.AssertExitCode(0),
		},
		Depends: []string{"Run Echo Container"},
	}

	// Remove sleep container
	removeSleep := &iapetus.Task{
		Name:    "Remove Sleep Container",
		Command: "docker",
		Args:    []string{"rm", "-f", "sleep-docker"},
		Asserts: []func(*iapetus.Task) error{
			iapetus.AssertExitCode(0),
		},
		Depends: []string{"Run Sleep Container"},
	}

	// Remove images
	removeAlpine := iapetus.NewTask("Remove Alpine Image", 10*time.Second, nil).
		AddCommand("docker").
		AddArgs("rmi", "-f", image).
		AssertExitCode(0)
	removeAlpine.Depends = []string{"Remove Echo Container"}

	removeBusybox := iapetus.NewTask("Remove Busybox Image", 10*time.Second, nil).
		AddCommand("docker").
		AddArgs("rmi", "-f", busybox).
		AssertExitCode(0)
	removeBusybox.Depends = []string{"Remove Sleep Container"}

	// Summary task depends on both image removals
	summary := &iapetus.Task{
		Name:    "Summary",
		Command: "echo",
		Args:    []string{"All Docker CLI tasks completed!"},
		Asserts: []func(*iapetus.Task) error{
			iapetus.AssertExitCode(0),
		},
		Depends: []string{"Remove Alpine Image", "Remove Busybox Image"},
	}

	workflow.Steps = []iapetus.Task{
		*pullAlpine,
		*pullBusybox,
		*runEcho,
		*runSleep,
		*removeEcho,
		*removeSleep,
		*removeAlpine,
		*removeBusybox,
		*summary,
	}

	fmt.Println("--- Running Docker CLI Complex Workflow Example ---")
	err := workflow.Run()
	if err != nil {
		fmt.Println("Workflow failed:", err)
		os.Exit(1)
	}
	fmt.Println("Workflow completed successfully!")
}
