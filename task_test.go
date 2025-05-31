package iapetus_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/yindia/iapetus"
)

func TestIntegrationTest_Run(t *testing.T) {
	test := &iapetus.Task{
		Command: "echo",
		Args:    []string{"Hello, World!"},
		Asserts: []func(*iapetus.Task) error{
			iapetus.AssertOutputEquals("Hello, World!\n"),
		},
	}
	test.SetBackend("bash")
	iapetus.RegisterBackend("bash", &iapetus.BashBackend{})

	err := test.Run()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if test.Actual.Output != "Hello, World!\n" {
		t.Errorf("expected output 'Hello, World!', got %v", test.Actual.Output)
	}
}

func TestIntegrationTest_AddAssertion(t *testing.T) {
	test := &iapetus.Task{}
	assertion := func(i *iapetus.Task) error {
		return nil
	}

	test.AddAssertion(assertion)

	if len(test.Asserts) != 1 {
		t.Errorf("expected 1 assertion, got %d", len(test.Asserts))
	}
}

func TestIntegrationTest_AddMultipleAssertions(t *testing.T) {
	test := &iapetus.Task{}
	assertion1 := func(i *iapetus.Task) error {
		return nil
	}
	assertion2 := func(i *iapetus.Task) error {
		return errors.New("failed assertion")
	}

	test.AddAssertion(assertion1)
	test.AddAssertion(assertion2)

	if len(test.Asserts) != 2 {
		t.Errorf("expected 2 assertions, got %d", len(test.Asserts))
	}
}

func TestIntegrationTest_RunCommandError(t *testing.T) {
	test := &iapetus.Task{
		Command: "invalid_command",
		Asserts: []func(*iapetus.Task) error{
			iapetus.AssertExitCode(1),
		},
	}

	err := test.Run()
	if err == nil {
		t.Fatalf("expected error for invalid command, got nil")
	}
}

func TestTask_AssertExitCode(t *testing.T) {
	task := iapetus.NewTask("test", 0, nil).AssertExitCode(0)
	task.Actual.ExitCode = 0
	if err := iapetus.RunAssertions(task); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	task.Actual.ExitCode = 1
	if err := iapetus.RunAssertions(task); err == nil {
		t.Errorf("expected error for wrong exit code, got nil")
	}
}

func TestTask_AssertOutputContains(t *testing.T) {
	task := iapetus.NewTask("test", 0, nil).AssertOutputContains("foo")
	task.Actual.Output = "hello foo bar"
	if err := iapetus.RunAssertions(task); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	task.Actual.Output = "hello bar"
	if err := iapetus.RunAssertions(task); err == nil {
		t.Errorf("expected error for missing substring, got nil")
	}
}

func TestTask_AssertOutputEquals(t *testing.T) {
	task := iapetus.NewTask("test", 0, nil).AssertOutputEquals("foo")
	task.Actual.Output = "foo"
	if err := iapetus.RunAssertions(task); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	task.Actual.Output = "bar"
	if err := iapetus.RunAssertions(task); err == nil {
		t.Errorf("expected error for output mismatch, got nil")
	}
}

func TestTask_AssertOutputMatchesRegexp(t *testing.T) {
	task := iapetus.NewTask("test", 0, nil).AssertOutputMatchesRegexp(`foo\d+`)
	task.Actual.Output = "foo123"
	if err := iapetus.RunAssertions(task); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	task.Actual.Output = "bar"
	if err := iapetus.RunAssertions(task); err == nil {
		t.Errorf("expected error for regex mismatch, got nil")
	}
}

func TestTask_ExpectDSL(t *testing.T) {
	task := iapetus.NewTask("test", 0, nil).
		Expect().
		ExitCode(0).
		OutputContains("foo").
		Done()
	task.Actual.ExitCode = 0
	task.Actual.Output = "foo bar"
	if err := iapetus.RunAssertions(task); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	task.Actual.ExitCode = 1
	if err := iapetus.RunAssertions(task); err == nil {
		t.Errorf("expected error for wrong exit code, got nil")
	}
}

func TestTask_EnvVars(t *testing.T) {
	t.Run("only EnvMap", func(t *testing.T) {
		b := &iapetus.BashBackend{}
		task := &iapetus.Task{
			Command: "sh",
			Args:    []string{"-c", "echo $FOO"},
			EnvMap:  map[string]string{"FOO": "bar"},
			Asserts: []func(*iapetus.Task) error{iapetus.AssertOutputEquals("bar\n")},
		}
		task.SetBackend("bash")
		iapetus.RegisterBackend("bash", b)
		if err := task.Run(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("EnvMap wins", func(t *testing.T) {
		b := &iapetus.BashBackend{}
		task := &iapetus.Task{
			Command: "sh",
			Args:    []string{"-c", "echo $FOO"},
			EnvMap:  map[string]string{"FOO": "baz"},
			Asserts: []func(*iapetus.Task) error{iapetus.AssertOutputEquals("baz\n")},
		}
		task.SetBackend("bash")
		iapetus.RegisterBackend("bash", b)
		if err := task.Run(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("no envs, default", func(t *testing.T) {
		b := &iapetus.BashBackend{}
		task := &iapetus.Task{
			Command: "sh",
			Args:    []string{"-c", "echo $FOO"},
			Asserts: []func(*iapetus.Task) error{iapetus.AssertOutputEquals("\n")},
		}
		task.SetBackend("bash")
		iapetus.RegisterBackend("bash", b)
		if err := task.Run(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

// --- Additional tests for Task backend/logger/env/validation ---

type testBackend struct {
	called      *bool
	fail        bool
	validateErr error
}

func (b *testBackend) RunTask(task *iapetus.Task) error {
	if b.called != nil {
		*b.called = true
	}
	if b.fail {
		return fmt.Errorf("backend run error")
	}
	return nil
}
func (b *testBackend) ValidateTask(task *iapetus.Task) error {
	return b.validateErr
}
func (b *testBackend) GetName() string {
	return "test"
}
func (b *testBackend) GetStatus() string {
	return "test"
}

func TestTask_DefaultsAndValidation(t *testing.T) {
	t.Run("defaults backend/logger/env", func(t *testing.T) {
		called := false
		iapetus.RegisterBackend("test", &testBackend{called: &called})
		task := &iapetus.Task{Command: "echo", Backend: "test"}
		if err := task.Run(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !called {
			t.Errorf("expected backend RunTask to be called")
		}
		if task.Logger() == nil {
			t.Errorf("expected logger to be set")
		}
		if task.EnvMap == nil {
			t.Errorf("expected EnvMap to be initialized")
		}
	})

	t.Run("missing command", func(t *testing.T) {
		task := &iapetus.Task{Backend: "test"}
		err := task.Run()
		if err == nil || err.Error() != "task task-: command is required" && !contains(err.Error(), "command is required") {
			t.Errorf("expected error for missing command, got %v", err)
		}
	})

	t.Run("validate error", func(t *testing.T) {
		iapetus.RegisterBackend("test-validate", &testBackend{validateErr: fmt.Errorf("bad task")})
		task := &iapetus.Task{Command: "echo", Backend: "test-validate"}
		err := task.Run()
		if err == nil || err.Error() != "bad task" {
			t.Errorf("expected validation error, got %v", err)
		}
	})

	t.Run("backend run error and retry", func(t *testing.T) {
		iapetus.RegisterBackend("test-fail", &testBackend{fail: true})
		task := &iapetus.Task{Command: "echo", Backend: "test-fail", Retries: 2}
		err := task.Run()
		if err == nil || !contains(err.Error(), "failed after 2 attempts") {
			t.Errorf("expected retry error, got %v", err)
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (contains(s[1:], substr) || contains(s[:len(s)-1], substr))))
}
