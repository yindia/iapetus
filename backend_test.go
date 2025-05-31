package iapetus

import (
	"errors"
	"os/exec"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

type mockBackend struct {
	name   string
	status string
	called *bool
}

func (m *mockBackend) RunTask(task *Task) error {
	if m.called != nil {
		*m.called = true
	}
	return nil
}
func (m *mockBackend) ValidateTask(task *Task) error { return nil }
func (m *mockBackend) GetName() string               { return m.name }
func (m *mockBackend) GetStatus() string             { return m.status }

func TestRegisterAndGetBackend(t *testing.T) {
	mb := &mockBackend{name: "mock", status: "available"}
	RegisterBackend("mock", mb)
	b := GetBackend("mock")
	if b == nil {
		t.Fatal("expected backend to be registered and retrievable")
	}
	if b.GetName() != "mock" {
		t.Errorf("expected GetName to return 'mock', got %q", b.GetName())
	}
	if b.GetStatus() != "available" {
		t.Errorf("expected GetStatus to return 'available', got %q", b.GetStatus())
	}
}

func TestGetBackend_NotFound(t *testing.T) {
	b := GetBackend("doesnotexist")
	if b != nil {
		t.Errorf("expected nil for unregistered backend, got %v", b)
	}
}

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
	task := NewTask("test", 0, zap.NewNop())
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
	task := NewTask("test", 0, zap.NewNop())
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
	task := NewTask("test", 5*time.Second, zap.NewNop())
	task.Image = "alpine"
	task.Command = "echo"
	task.Args = []string{"hello"}
	task.Asserts = []func(*Task) error{AssertOutputContains("hello")}
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
	task := NewTask("test", 5*time.Second, zap.NewNop())
	task.Image = "alpine"
	task.Command = "sh"
	task.Args = []string{"-c", "echo $FOO"}
	task.EnvMap = map[string]string{"FOO": "bar"}
	task.Asserts = []func(*Task) error{AssertOutputEquals("bar\n")}
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
	task := NewTask("test", 5*time.Second, zap.NewNop())
	task.Image = "alpine"
	task.Command = "echo"
	task.Args = []string{"foo"}
	task.Asserts = []func(*Task) error{func(t *Task) error { return errors.New("fail") }}
	if err := b.RunTask(task); err == nil || !strings.Contains(err.Error(), "fail") {
		t.Errorf("expected assertion error, got %v", err)
	}
}

func TestBashBackend_ValidateTask(t *testing.T) {
	b := &BashBackend{}
	task := NewTask("test", 0, zap.NewNop())
	task.Command = "echo"
	if err := b.ValidateTask(task); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestBashBackend_RunTask_Success(t *testing.T) {
	b := &BashBackend{}
	task := NewTask("test", 2*time.Second, zap.NewNop())
	task.Command = "echo"
	task.Args = []string{"hello"}
	task.Asserts = []func(*Task) error{AssertOutputContains("hello")}
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

func TestBashBackend_RunTask_EnvMap(t *testing.T) {
	b := &BashBackend{}
	task := NewTask("test", 2*time.Second, zap.NewNop())
	task.Command = "sh"
	task.Args = []string{"-c", "echo $FOO"}
	task.EnvMap = map[string]string{"FOO": "bar"}
	task.Asserts = []func(*Task) error{AssertOutputEquals("bar\n")}
	if err := b.RunTask(task); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if task.Actual.Output != "bar\n" {
		t.Errorf("expected output 'bar\n', got %q", task.Actual.Output)
	}
}

func TestBashBackend_RunTask_Timeout(t *testing.T) {
	b := &BashBackend{}
	task := NewTask("test", 500*time.Millisecond, zap.NewNop())
	task.Command = "sleep"
	task.Args = []string{"2"}
	if err := b.RunTask(task); err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Errorf("expected timeout error, got %v", err)
	}
}

func TestBashBackend_RunTask_AssertionFail(t *testing.T) {
	b := &BashBackend{}
	task := NewTask("test", 2*time.Second, zap.NewNop())
	task.Command = "echo"
	task.Args = []string{"foo"}
	task.Asserts = []func(*Task) error{func(t *Task) error { return errors.New("fail") }}
	if err := b.RunTask(task); err == nil || !strings.Contains(err.Error(), "fail") {
		t.Errorf("expected assertion error, got %v", err)
	}
}

func kubectlAvailable() bool {
	_, err := exec.LookPath("kubectl")
	if err != nil {
		return false
	}
	cmd := exec.Command("kubectl", "version", "--client")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func TestKubernetesBackend_ValidateTask(t *testing.T) {
	b := &KubernetesBackend{}
	task := NewTask("test", 0, zap.NewNop())
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

func TestKubernetesBackend_RunTask(t *testing.T) {
	if !kubectlAvailable() {
		t.Skip("kubectl not available")
	}
	b := &KubernetesBackend{}
	task := NewTask("k8s-echo", 5*time.Second, zap.NewNop())
	task.Image = "alpine"
	task.Command = "echo"
	task.Args = []string{"hello-from-k8s"}
	task.Asserts = []func(*Task) error{AssertOutputContains("hello-from-k8s")}
	if err := b.RunTask(task); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(task.Actual.Output, "hello-from-k8s") {
		t.Errorf("expected output to contain 'hello-from-k8s', got %q", task.Actual.Output)
	}
}
