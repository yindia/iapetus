package iapetus

import (
	"errors"
	"testing"
)

func TestIntegrationTest_Run(t *testing.T) {
	test := &Task{
		Command: "echo",
		Args:    []string{"Hello, World!"},
		Asserts: []func(*Task) error{
			AssertOutputEquals("Hello, World!\n"),
		},
	}

	err := test.Run()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if test.Actual.Output != "Hello, World!\n" {
		t.Errorf("expected output 'Hello, World!', got %v", test.Actual.Output)
	}
}

func TestIntegrationTest_AddAssertion(t *testing.T) {
	test := &Task{}
	assertion := func(i *Task) error {
		return nil
	}

	test.AddAssertion(assertion)

	if len(test.Asserts) != 1 {
		t.Errorf("expected 1 assertion, got %d", len(test.Asserts))
	}
}

func TestIntegrationTest_AddMultipleAssertions(t *testing.T) {
	test := &Task{}
	assertion1 := func(i *Task) error {
		return nil
	}
	assertion2 := func(i *Task) error {
		return errors.New("failed assertion")
	}

	test.AddAssertion(assertion1)
	test.AddAssertion(assertion2)

	if len(test.Asserts) != 2 {
		t.Errorf("expected 2 assertions, got %d", len(test.Asserts))
	}
}

func TestIntegrationTest_RunCommandError(t *testing.T) {
	test := &Task{
		Command: "invalid_command",
		Asserts: []func(*Task) error{
			AssertExitCode(1),
		},
	}

	err := test.Run()
	if err == nil {
		t.Fatalf("expected error for invalid command, got nil")
	}
}

func TestTask_AssertExitCode(t *testing.T) {
	task := NewTask("test", 0, nil).AssertExitCode(0)
	task.Actual.ExitCode = 0
	if err := RunAssertions(task); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	task.Actual.ExitCode = 1
	if err := RunAssertions(task); err == nil {
		t.Errorf("expected error for wrong exit code, got nil")
	}
}

func TestTask_AssertOutputContains(t *testing.T) {
	task := NewTask("test", 0, nil).AssertOutputContains("foo")
	task.Actual.Output = "hello foo bar"
	if err := RunAssertions(task); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	task.Actual.Output = "hello bar"
	if err := RunAssertions(task); err == nil {
		t.Errorf("expected error for missing substring, got nil")
	}
}

func TestTask_AssertOutputEquals(t *testing.T) {
	task := NewTask("test", 0, nil).AssertOutputEquals("foo")
	task.Actual.Output = "foo"
	if err := RunAssertions(task); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	task.Actual.Output = "bar"
	if err := RunAssertions(task); err == nil {
		t.Errorf("expected error for output mismatch, got nil")
	}
}

func TestTask_AssertOutputMatchesRegexp(t *testing.T) {
	task := NewTask("test", 0, nil).AssertOutputMatchesRegexp(`foo\d+`)
	task.Actual.Output = "foo123"
	if err := RunAssertions(task); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	task.Actual.Output = "bar"
	if err := RunAssertions(task); err == nil {
		t.Errorf("expected error for regex mismatch, got nil")
	}
}

func TestTask_ExpectDSL(t *testing.T) {
	task := NewTask("test", 0, nil).
		Expect().
		ExitCode(0).
		OutputContains("foo").
		Done()
	task.Actual.ExitCode = 0
	task.Actual.Output = "foo bar"
	if err := RunAssertions(task); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	task.Actual.ExitCode = 1
	if err := RunAssertions(task); err == nil {
		t.Errorf("expected error for wrong exit code, got nil")
	}
}
