package iapetus

import (
	"os"
	"testing"
)

func TestLoadWorkflowFromYAML_Success(t *testing.T) {
	yamlContent := `
name: test-wf
backend: bash
env_map:
  FOO: bar
steps:
  - name: step1
    command: echo
    args: ["hello"]
    timeout: 1s
    backend: bash
    env_map:
      BAR: baz
    raw_asserts:
      - exit_code: 0
      - output_contains: hello
      - output_equals: "hello\n"
      - output_json_equals: '{"foo": 1}'
      - output_matches_regexp: '^hello.*$'
      - output_json_equals: '{"foo": 1}'
        skip_json_nodes: ["foo.bar"]
  - name: step2
    command: echo
    args: ["world"]
    depends: [step1]
    raw_asserts:
      - output_equals: "world\n"
`
	f, err := os.CreateTemp("", "iapetus_yaml_test_*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString(yamlContent); err != nil {
		t.Fatalf("failed to write yaml: %v", err)
	}
	f.Close()

	wf, err := LoadWorkflowFromYAML(f.Name())
	if err != nil {
		t.Fatalf("LoadWorkflowFromYAML failed: %v", err)
	}
	if wf.Name != "test-wf" {
		t.Errorf("expected workflow name 'test-wf', got %q", wf.Name)
	}
	if wf.Backend != "bash" {
		t.Errorf("expected backend 'bash', got %q", wf.Backend)
	}
	if wf.EnvMap["FOO"] != "bar" {
		t.Errorf("expected env_map FOO=bar, got %v", wf.EnvMap)
	}
	if len(wf.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(wf.Steps))
	}
	step1 := wf.Steps[0]
	if step1.Name != "step1" || step1.Command != "echo" {
		t.Errorf("unexpected step1: %+v", step1)
	}
	if step1.Timeout.Seconds() != 1 {
		t.Errorf("expected timeout 1s, got %v", step1.Timeout)
	}
	if step1.EnvMap["BAR"] != "baz" {
		t.Errorf("expected env_map BAR=baz, got %v", step1.EnvMap)
	}
	if len(step1.Asserts) != 6 {
		t.Errorf("expected 6 assertions, got %d", len(step1.Asserts))
	}
}

func TestLoadWorkflowFromYAML_FileNotFound(t *testing.T) {
	_, err := LoadWorkflowFromYAML("nonexistent.yaml")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestLoadWorkflowFromYAML_InvalidYAML(t *testing.T) {
	f, err := os.CreateTemp("", "iapetus_yaml_test_invalid_*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString("not: [valid: yaml"); err != nil {
		t.Fatalf("failed to write yaml: %v", err)
	}
	f.Close()
	_, err = LoadWorkflowFromYAML(f.Name())
	if err == nil {
		t.Error("expected error for invalid yaml, got nil")
	}
}
