package docker

import (
	"fmt"
	"os/exec"

	"github.com/yindia/iapetus"
)

func init() {
	iapetus.RegisterBackend("docker", &DockerBackend{})
}

type DockerBackend struct{}

// ValidateTask checks if the task is valid for Docker execution.
func (d *DockerBackend) ValidateTask(task *iapetus.Task) error {
	if task.Image == "" {
		return fmt.Errorf("docker backend requires task.Image to be set")
	}
	if task.Command == "" {
		return fmt.Errorf("docker backend requires task.Command to be set")
	}
	return nil
}

// RunTask executes the task in a Docker container.
func (d *DockerBackend) RunTask(task *iapetus.Task) error {
	if err := d.ValidateTask(task); err != nil {
		return err
	}

	dockerArgs := []string{"run", "--rm"}
	if task.WorkingDir != "" {
		dockerArgs = append(dockerArgs, "-w", task.WorkingDir)
	}
	for k, v := range task.EnvMap {
		dockerArgs = append(dockerArgs, "-e", fmt.Sprintf("%s=%s", k, v))
	}
	dockerArgs = append(dockerArgs, task.Image)
	dockerArgs = append(dockerArgs, task.Command)
	dockerArgs = append(dockerArgs, task.Args...)

	cmd := exec.Command("docker", dockerArgs...)
	output, err := cmd.CombinedOutput()
	task.Actual.Output = string(output)
	task.Actual.ExitCode = 0
	if err != nil {
		task.Actual.Error = err.Error()
		if exitErr, ok := err.(*exec.ExitError); ok {
			task.Actual.ExitCode = exitErr.ExitCode()
		} else {
			task.Actual.ExitCode = 1
		}
		return fmt.Errorf("docker run failed: %w\nOutput: %s", err, output)
	}
	// Run assertions and propagate errors
	err = iapetus.RunAssertions(task)
	if err != nil {
		return err
	}
	return nil
}
