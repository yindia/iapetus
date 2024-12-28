package iapetus

import (
	"errors"
	"testing"
)

func TestIntegrationTest_Run(t *testing.T) {
	test := &Step{
		Command: "echo",
		Args:    []string{"Hello, World!"},
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
	test := &Step{}
	assertion := func(i *Step) error {
		return nil
	}

	test.AddAssertion(assertion)

	if len(test.Asserts) != 1 {
		t.Errorf("expected 1 assertion, got %d", len(test.Asserts))
	}
}

func TestIntegrationTest_AddMultipleAssertions(t *testing.T) {
	test := &Step{}
	assertion1 := func(i *Step) error {
		return nil
	}
	assertion2 := func(i *Step) error {
		return errors.New("failed assertion")
	}

	test.AddAssertion(assertion1)
	test.AddAssertion(assertion2)

	if len(test.Asserts) != 2 {
		t.Errorf("expected 2 assertions, got %d", len(test.Asserts))
	}
}

func TestIntegrationTest_RunCommandError(t *testing.T) {
	test := &Step{
		Command: "invalid_command",

		Expected: Output{
			ExitCode: 1,
		},
		Asserts: []func(*Step) error{
			AssertByExitCode,
		},
	}

	err := test.Run()
	if err == nil {
		t.Fatalf("expected error for invalid command, got nil")
	}
}