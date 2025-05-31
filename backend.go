package iapetus

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"go.uber.org/zap"
)

func init() {
	RegisterBackend("bash", &BashBackend{})
	RegisterBackend("docker", &DockerBackend{})
}

// Backend is the interface for task execution plugins.
type Backend interface {
	RunTask(task *Task) error
	ValidateTask(task *Task) error
	GetName() string
	GetStatus() string
}

// backendRegistry holds all registered backends by name.
var backendRegistry = map[string]Backend{}

// RegisterBackend registers a backend plugin by name.
// Plugin authors: call this in your plugin's init() function.
func RegisterBackend(name string, backend Backend) {
	backendRegistry[name] = backend
}

// GetBackend retrieves a backend by name, or nil if not found.
func GetBackend(name string) Backend {
	if b, ok := backendRegistry[name]; ok {
		return b
	}
	return nil
}

type BashBackend struct{}

func (b *BashBackend) ValidateTask(t *Task) error {
	return nil
}

func (b *BashBackend) RunTask(t *Task) error {
	t.EnsureDefaults()
	ctx, cancel := context.WithTimeout(context.Background(), t.Timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, t.Command, t.Args...)

	// Merge environment variables: os.Environ + t.Env + t.EnvMap (EnvMap takes precedence)
	envMap := map[string]string{}
	for _, kv := range os.Environ() {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	for k, v := range t.EnvMap {
		envMap[k] = v
	}
	finalEnv := make([]string, 0, len(envMap))
	for k, v := range envMap {
		finalEnv = append(finalEnv, k+"="+v)
	}
	cmd.Env = finalEnv

	if t.WorkingDir != "" {
		cmd.Dir = t.WorkingDir
	}
	t.Logger().Debug("Command", zap.String("cmd", t.Command+" "+strings.Join(t.Args, " ")))
	output, err := cmd.CombinedOutput()
	t.Actual.Output = string(output)
	t.Actual.ExitCode = GetExitCode(err)
	if err != nil {
		t.Actual.Error = err.Error()
		if ctx.Err() == context.DeadlineExceeded {
			t.Logger().Error("Task timed out", zap.String("task", t.Name), zap.Duration("timeout", t.Timeout))
			return fmt.Errorf("task %s timed out after %v", t.Name, t.Timeout)
		}
		t.Logger().Error("Error executing task", zap.String("task", t.Name), zap.Error(err))
	}
	// Use RunAssertions to aggregate all assertion errors
	err = RunAssertions(t)
	if err != nil {
		t.Logger().Error("Assertion(s) failed", zap.String("task", t.Name), zap.Error(err))
		return err
	}
	return nil
}

func (b *BashBackend) GetName() string {
	return "bash"
}

func (b *BashBackend) GetStatus() string {
	return "available"
}

type DockerBackend struct{}

// ValidateTask checks if the task is valid for Docker execution.
func (d *DockerBackend) ValidateTask(task *Task) error {
	if task.Image == "" {
		return fmt.Errorf("docker backend requires task.Image to be set")
	}
	if task.Command == "" {
		return fmt.Errorf("docker backend requires task.Command to be set")
	}
	return nil
}

// RunTask executes the task in a Docker container.
func (d *DockerBackend) RunTask(task *Task) error {
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
	err = RunAssertions(task)
	if err != nil {
		return err
	}
	return nil
}

func (d *DockerBackend) GetName() string {
	return "docker"
}

func (d *DockerBackend) GetStatus() string {
	if _, err := exec.LookPath("docker"); err == nil {
		return "available"
	}
	return "unavailable"
}
