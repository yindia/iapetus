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
//     raw_asserts:
//   - exit_code: 0
//   - output_contains: hello
//   - output_equals: "hello\n"
//   - output_json_equals: '{"foo": 1}'
//   - output_matches_regexp: '^hello.*$'
//   - output_json_equals: '{"foo": 1}'
//     skip_json_nodes: ["foo.bar"]
//   - name: step2
//     command: echo
//     args: ["world"]
//     depends: [step1]
//     raw_asserts:
//   - output_equals: "world\n"
//
// Note: Only fields that can be represented in YAML (strings, ints, slices, maps, etc.)
// are supported. Assertions (functions) must be added programmatically after loading using raw_asserts.
//
// Usage:
//
//	wf, err := iapetus.LoadWorkflowFromYAML("workflow.yaml")
//	if err != nil { log.Fatal(err) }
//	err = wf.Run()
package iapetus

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// DefaultRetryDelay is the default delay between retries if not specified per task.
var DefaultRetryDelay time.Duration = 1 * time.Second

// assertionYAML is a helper struct for parsing assertions from YAML
// Supports all built-in assertion types.
type assertionYAML struct {
	ExitCode            *int     `yaml:"exit_code,omitempty"`
	OutputEquals        *string  `yaml:"output_equals,omitempty"`
	OutputContains      *string  `yaml:"output_contains,omitempty"`
	OutputJsonEquals    *string  `yaml:"output_json_equals,omitempty"`
	OutputMatchesRegexp *string  `yaml:"output_matches_regexp,omitempty"`
	SkipJsonNodes       []string `yaml:"skip_json_nodes,omitempty"`
}

type taskYAML struct {
	Name       string            `yaml:"name"`
	Command    string            `yaml:"command"`
	Args       []string          `yaml:"args,omitempty"`
	Timeout    string            `yaml:"timeout,omitempty"`
	Retries    int               `yaml:"retries,omitempty"`
	RetryDelay string            `yaml:"retry_delay,omitempty"` // Delay between retries (e.g. "2s"). Defaults to 1s if not set.
	Depends    []string          `yaml:"depends,omitempty"`
	EnvMap     map[string]string `yaml:"env_map,omitempty"`
	Image      string            `yaml:"image,omitempty"`
	Backend    string            `yaml:"backend,omitempty"`
	RawAsserts []assertionYAML   `yaml:"raw_asserts,omitempty"`
}

type workflowYAML struct {
	Name    string            `yaml:"name"`
	Backend string            `yaml:"backend,omitempty"`
	EnvMap  map[string]string `yaml:"env_map,omitempty"`
	Steps   []taskYAML        `yaml:"steps"`
}

// LoadWorkflowFromYAML loads a Workflow from a YAML file.
//
// All fields except assertions (Asserts) are loaded from YAML.
// Assertions are specified in raw_asserts and converted after loading.
//
// Example:
//
//	wf, err := iapetus.LoadWorkflowFromYAML("workflow.yaml")
//	if err != nil { log.Fatal(err) }
//	err = wf.Run()
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
		if t.RetryDelay != "" {
			dur, err := time.ParseDuration(t.RetryDelay)
			if err != nil {
				return nil, fmt.Errorf("invalid retry_delay for task %s: %w", t.Name, err)
			}
			task.RetryDelay = dur
		} else {
			task.RetryDelay = DefaultRetryDelay
		}
		for _, a := range t.RawAsserts {
			if a.ExitCode != nil {
				task.Asserts = append(task.Asserts, AssertExitCode(*a.ExitCode))
			}
			if a.OutputEquals != nil {
				task.Asserts = append(task.Asserts, AssertOutputEquals(*a.OutputEquals))
			}
			if a.OutputContains != nil {
				task.Asserts = append(task.Asserts, AssertOutputContains(*a.OutputContains))
			}
			if a.OutputJsonEquals != nil {
				task.Asserts = append(task.Asserts, AssertOutputJsonEquals(*a.OutputJsonEquals, a.SkipJsonNodes...))
			}
			if a.OutputMatchesRegexp != nil {
				task.Asserts = append(task.Asserts, AssertOutputMatchesRegexp(*a.OutputMatchesRegexp))
			}
		}
		wf.AddTask(task)
	}
	return wf, nil
}
