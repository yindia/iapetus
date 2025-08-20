// Package iapetus provides a plugin-based workflow engine for automating and testing command-line tasks.
//
// The backend subpackage defines the Backend interface and built-in backends (bash, docker).
// Plugin authors can implement custom backends by satisfying the Backend interface and registering them with RegisterBackend.
//
// # Backend Plugin System
//
// - Implement the Backend interface for your environment (e.g., Kubernetes, SSH, etc).
// - Register your backend in an init() function using RegisterBackend.
// - Tasks and workflows can select a backend by name.
//
// # Built-in Backends
//
// - BashBackend: Runs tasks as local shell commands.
// - DockerBackend: Runs tasks in Docker containers (requires task.Image).
//
// See the documentation for more details and examples.
package iapetus

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"go.uber.org/zap"
)

// init registers the built-in backends (bash, docker) at startup.
func init() {
	RegisterBackend("bash", &BashBackend{})
	RegisterBackend("docker", &DockerBackend{})
	RegisterBackend("kubernetes", &KubernetesBackend{})
}

// Backend is the interface for task execution plugins.
//
// Implement this interface to add a new backend (e.g., for Docker, Kubernetes, SSH, etc).
// Register your backend with RegisterBackend.
type Backend interface {
	// RunTask executes the given task and populates its Actual fields.
	RunTask(task *Task) error
	// ValidateTask checks if the task is valid for this backend.
	ValidateTask(task *Task) error
	// GetName returns the backend's name (for registry and diagnostics).
	GetName() string
	// GetStatus returns a status string (e.g., "available", "unavailable").
	GetStatus() string
}

// backendRegistry holds all registered backends by name.
var backendRegistry = map[string]Backend{}

// RegisterBackend registers a backend plugin by name.
//
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

// BashBackend runs tasks as local shell commands.
type BashBackend struct{}

// ValidateTask always returns nil for BashBackend (all commands allowed).
func (b *BashBackend) ValidateTask(t *Task) error {
	return nil
}

// RunTask executes the task as a local shell command.
// Merges environment variables from os.Environ and task.EnvMap.
// Populates task.Actual.Output, ExitCode, and Error.
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

	var outputBuffer bytes.Buffer
	if t.Stdout != nil {
		cmd.Stdout = t.Stdout
	} else {
		cmd.Stdout = &outputBuffer
	}
	if t.Stderr != nil {
		cmd.Stderr = t.Stderr
	} else {
		cmd.Stderr = &outputBuffer
	}

	err := cmd.Run()
	t.Actual.Output = outputBuffer.String()
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

// GetName returns the backend name ("bash").
func (b *BashBackend) GetName() string {
	return "bash"
}

// GetStatus returns "available" for BashBackend.
func (b *BashBackend) GetStatus() string {
	return "available"
}

// DockerBackend runs tasks in Docker containers.
type DockerBackend struct{}

// ValidateTask checks if the task is valid for Docker execution.
// Requires task.Image and task.Command to be set.
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
// Passes environment variables, working directory, and arguments to the container.
// Populates task.Actual.Output, ExitCode, and Error.
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

// GetName returns the backend name ("docker").
func (d *DockerBackend) GetName() string {
	return "docker"
}

// GetStatus returns "available" if Docker is installed, else "unavailable".
func (d *DockerBackend) GetStatus() string {
	if _, err := exec.LookPath("docker"); err == nil {
		return "available"
	}
	return "unavailable"
}

// KubernetesBackend runs tasks in Kubernetes pods using kubectl.
type KubernetesBackend struct{}

// ValidateTask checks if the task is valid for Kubernetes execution.
func (k *KubernetesBackend) ValidateTask(task *Task) error {
	if task.Image == "" {
		return fmt.Errorf("kubernetes backend requires task.Image to be set")
	}
	if task.Command == "" {
		return fmt.Errorf("kubernetes backend requires task.Command or task.Script to be set")
	}
	return nil
}

// RunTask executes the task in a Kubernetes pod using kubectl.
// For demo: uses 'kubectl run' and waits for completion.
// Limitations: Only works if kubectl is installed and configured. Env vars are not injected via --env for simplicity.
func (k *KubernetesBackend) RunTask(task *Task) error {
	if err := k.ValidateTask(task); err != nil {
		return err
	}

	// Build the command to run in the pod
	var cmdStr string
	cmdStr = task.Command
	if len(task.Args) > 0 {
		cmdStr += " " + strings.Join(task.Args, " ")
	}

	// Generate a unique pod name
	podName := fmt.Sprintf("iapetus-%s-%d", strings.ToLower(task.Name), os.Getpid())

	// Build kubectl args
	kubectlArgs := []string{
		"run", podName,
		"--image", task.Image,
		"--restart", "Never",
		"--rm", // auto-delete pod after completion
		"--attach",
		"--command", "--",
		"sh", "-c", cmdStr,
	}
	cmd := exec.Command("kubectl", kubectlArgs...)

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
		return fmt.Errorf("kubectl run failed: %w\nOutput: %s", err, output)
	}
	// Run assertions and propagate errors
	err = RunAssertions(task)
	if err != nil {
		return err
	}
	return nil
}

func (k *KubernetesBackend) GetName() string { return "kubernetes" }
func (k *KubernetesBackend) GetStatus() string {
	if _, err := exec.LookPath("kubectl"); err == nil {
		return "available"
	}
	return "unavailable"
}
