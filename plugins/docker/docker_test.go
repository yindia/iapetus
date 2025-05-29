package docker

import (
	"errors"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/yindia/iapetus"
	"go.uber.org/zap"
)

func dockerAvailable() bool {
	_, err := exec.LookPath("docker")
	if err != nil {
		return false
	}
	cmd := exec.Command("docker", "version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func TestDockerBackend_ValidateTask(t *testing.T) {
	b := &DockerBackend{}
	task := iapetus.NewTask("test", 0, zap.NewNop())
	task.Image = "alpine"
	task.Command = "echo"
	if err := b.ValidateTask(task); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	task.Image = ""
	if err := b.ValidateTask(task); err == nil || !strings.Contains(err.Error(), "Image") {
		t.Errorf("expected image error, got %v", err)
	}
	task.Image = "alpine"
	task.Command = ""
	if err := b.ValidateTask(task); err == nil || !strings.Contains(err.Error(), "Command") {
		t.Errorf("expected command error, got %v", err)
	}
}

func TestDockerBackend_RunTask_ValidateFail(t *testing.T) {
	b := &DockerBackend{}
	task := iapetus.NewTask("test", 0, zap.NewNop())
	task.Image = ""
	task.Command = "echo"
	if err := b.RunTask(task); err == nil || !strings.Contains(err.Error(), "Image") {
		t.Errorf("expected image error, got %v", err)
	}
}

func TestDockerBackend_RunTask_Success(t *testing.T) {
	if !dockerAvailable() {
		t.Skip("docker not available")
	}
	b := &DockerBackend{}
	task := iapetus.NewTask("test", 5*time.Second, zap.NewNop())
	task.Image = "alpine"
	task.Command = "echo"
	task.Args = []string{"hello"}
	task.Asserts = []func(*iapetus.Task) error{iapetus.AssertOutputContains("hello")}
	if err := b.RunTask(task); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(task.Actual.Output, "hello") {
		t.Errorf("expected output to contain 'hello', got %q", task.Actual.Output)
	}
	if task.Actual.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", task.Actual.ExitCode)
	}
}

func TestDockerBackend_RunTask_EnvMap(t *testing.T) {
	if !dockerAvailable() {
		t.Skip("docker not available")
	}
	b := &DockerBackend{}
	task := iapetus.NewTask("test", 5*time.Second, zap.NewNop())
	task.Image = "alpine"
	task.Command = "sh"
	task.Args = []string{"-c", "echo $FOO"}
	task.EnvMap = map[string]string{"FOO": "bar"}
	task.Asserts = []func(*iapetus.Task) error{iapetus.AssertOutputEquals("bar\n")}
	if err := b.RunTask(task); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if task.Actual.Output != "bar\n" {
		t.Errorf("expected output 'bar\n', got %q", task.Actual.Output)
	}
}

func TestDockerBackend_RunTask_AssertionFail(t *testing.T) {
	if !dockerAvailable() {
		t.Skip("docker not available")
	}
	b := &DockerBackend{}
	task := iapetus.NewTask("test", 5*time.Second, zap.NewNop())
	task.Image = "alpine"
	task.Command = "echo"
	task.Args = []string{"foo"}
	task.Asserts = []func(*iapetus.Task) error{func(t *iapetus.Task) error { return errors.New("fail") }}
	if err := b.RunTask(task); err == nil || !strings.Contains(err.Error(), "fail") {
		t.Errorf("expected assertion error, got %v", err)
	}
}
