// yaml.go
//
// Example YAML schema:
//
// name: my-workflow
// backend: bash
// env_map:
//
//	FOO: bar
//
// steps:
//   - name: step1
//     command: echo
//     args: ["hello"]
//     timeout: 10s
//     backend: bash
//     env_map:
//     BAR: baz
//     asserts:
//   - exit_code: 0
//   - output_contains: hello
//   - name: step2
//     command: echo
//     args: ["world"]
//     depends: [step1]
//     asserts:
//   - output_equals: world\n
package iapetus

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// assertionYAML is a helper struct for parsing assertions from YAML
// (expand as needed for more assertion types)
type assertionYAML struct {
	ExitCode       *int    `yaml:"exit_code,omitempty"`
	OutputEquals   *string `yaml:"output_equals,omitempty"`
	OutputContains *string `yaml:"output_contains,omitempty"`
}

type taskYAML struct {
	Name    string            `yaml:"name"`
	Command string            `yaml:"command"`
	Args    []string          `yaml:"args,omitempty"`
	Timeout string            `yaml:"timeout,omitempty"`
	Retries int               `yaml:"retries,omitempty"`
	Depends []string          `yaml:"depends,omitempty"`
	EnvMap  map[string]string `yaml:"env_map,omitempty"`
	Image   string            `yaml:"image,omitempty"`
	Backend string            `yaml:"backend,omitempty"`
	Asserts []assertionYAML   `yaml:"asserts,omitempty"`
}

type workflowYAML struct {
	Name    string            `yaml:"name"`
	Backend string            `yaml:"backend,omitempty"`
	EnvMap  map[string]string `yaml:"env_map,omitempty"`
	Steps   []taskYAML        `yaml:"steps"`
}

// LoadWorkflowFromYAML loads a workflow from a YAML file.
func LoadWorkflowFromYAML(path string) (*Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}
	var wfY workflowYAML
	if err := yaml.Unmarshal(data, &wfY); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	wf := NewWorkflow(wfY.Name, nil)
	if wfY.Backend != "" {
		wf.Backend = wfY.Backend
	}
	if wfY.EnvMap != nil {
		wf.EnvMap = wfY.EnvMap
	}
	for _, t := range wfY.Steps {
		task := Task{
			Name:    t.Name,
			Command: t.Command,
			Args:    t.Args,
			Retries: t.Retries,
			Depends: t.Depends,
			EnvMap:  t.EnvMap,
			Image:   t.Image,
		}
		if t.Backend != "" {
			task.Backend = t.Backend
		}
		if t.Timeout != "" {
			dur, err := time.ParseDuration(t.Timeout)
			if err != nil {
				return nil, fmt.Errorf("invalid timeout for task %s: %w", t.Name, err)
			}
			task.Timeout = dur
		}
		// Parse assertions
		for _, a := range t.Asserts {
			if a.ExitCode != nil {
				task.Asserts = append(task.Asserts, AssertExitCode(*a.ExitCode))
			}
			if a.OutputEquals != nil {
				task.Asserts = append(task.Asserts, AssertOutputEquals(*a.OutputEquals))
			}
			if a.OutputContains != nil {
				task.Asserts = append(task.Asserts, AssertOutputContains(*a.OutputContains))
			}
		}
		wf.AddTask(task)
	}
	return wf, nil
}
