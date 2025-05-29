package bash

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/yindia/iapetus"
	"go.uber.org/zap"
)

func TestBashBackend_ValidateTask(t *testing.T) {
	b := &BashBackend{}
	task := iapetus.NewTask("test", 0, zap.NewNop())
	task.Command = "echo"
	if err := b.ValidateTask(task); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestBashBackend_RunTask_Success(t *testing.T) {
	b := &BashBackend{}
	task := iapetus.NewTask("test", 2*time.Second, zap.NewNop())
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

func TestBashBackend_RunTask_EnvMap(t *testing.T) {
	b := &BashBackend{}
	task := iapetus.NewTask("test", 2*time.Second, zap.NewNop())
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

func TestBashBackend_RunTask_Timeout(t *testing.T) {
	b := &BashBackend{}
	task := iapetus.NewTask("test", 500*time.Millisecond, zap.NewNop())
	task.Command = "sleep"
	task.Args = []string{"2"}
	if err := b.RunTask(task); err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Errorf("expected timeout error, got %v", err)
	}
}

func TestBashBackend_RunTask_AssertionFail(t *testing.T) {
	b := &BashBackend{}
	task := iapetus.NewTask("test", 2*time.Second, zap.NewNop())
	task.Command = "echo"
	task.Args = []string{"foo"}
	task.Asserts = []func(*iapetus.Task) error{func(t *iapetus.Task) error { return errors.New("fail") }}
	if err := b.RunTask(task); err == nil || !strings.Contains(err.Error(), "fail") {
		t.Errorf("expected assertion error, got %v", err)
	}
}
