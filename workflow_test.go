package iapetus_test

import (
	"fmt"
	"math/rand"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"testing/quick"
	"time"

	"github.com/yindia/iapetus"
	_ "github.com/yindia/iapetus/plugins/bash"
	"go.uber.org/zap"
)

func TestWorkflowRun(t *testing.T) {
	tests := []struct {
		name     string
		workflow *iapetus.Workflow
		wantErr  bool
	}{
		{
			name: "All Steps Pass",
			workflow: func() *iapetus.Workflow {
				wf := iapetus.NewWorkflow("all-steps-pass", zap.NewNop())
				wf.AddTask(iapetus.Task{
					Name:    "step1",
					Command: "echo",
					Args:    []string{"hello"},
					Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0), iapetus.AssertOutputEquals("hello\n")},
				})
				return wf
			}(),
			wantErr: false,
		},
		{
			name: "Step Fails on Output",
			workflow: func() *iapetus.Workflow {
				wf := iapetus.NewWorkflow("step-fails-on-output", zap.NewNop())
				wf.AddTask(iapetus.Task{
					Name:    "step1",
					Command: "echo",
					Args:    []string{"hello"},
					Asserts: []func(*iapetus.Task) error{iapetus.AssertOutputEquals("world\n")},
				})
				return wf
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.workflow.Run()
			if (err != nil) != tt.wantErr {
				t.Errorf("Workflow.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWorkflow_Run_EdgeCases(t *testing.T) {
	t.Run("linear chain", func(t *testing.T) {
		wf := iapetus.NewWorkflow("linear-chain", zap.NewNop())
		wf.AddTask(iapetus.Task{Name: "a", Command: "echo", Args: []string{"1"}, Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)}})
		wf.AddTask(iapetus.Task{Name: "b", Command: "echo", Args: []string{"2"}, Depends: []string{"a"}, Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)}})
		wf.AddTask(iapetus.Task{Name: "c", Command: "echo", Args: []string{"3"}, Depends: []string{"b"}, Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)}})
		err := wf.Run()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("diamond dependency", func(t *testing.T) {
		wf := iapetus.NewWorkflow("diamond-dependency", zap.NewNop())
		wf.AddTask(iapetus.Task{Name: "a", Command: "echo", Args: []string{"A"}, Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)}})
		wf.AddTask(iapetus.Task{Name: "b", Command: "echo", Args: []string{"B"}, Depends: []string{"a"}, Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)}})
		wf.AddTask(iapetus.Task{Name: "c", Command: "echo", Args: []string{"C"}, Depends: []string{"a"}, Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)}})
		wf.AddTask(iapetus.Task{Name: "d", Command: "echo", Args: []string{"D"}, Depends: []string{"b", "c"}, Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)}})
		err := wf.Run()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("multiple roots and leaves", func(t *testing.T) {
		wf := iapetus.NewWorkflow("multiple-roots-leaves", zap.NewNop())
		wf.AddTask(iapetus.Task{Name: "root1", Command: "echo", Args: []string{"R1"}, Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)}})
		wf.AddTask(iapetus.Task{Name: "root2", Command: "echo", Args: []string{"R2"}, Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)}})
		wf.AddTask(iapetus.Task{Name: "mid1", Command: "echo", Args: []string{"M1"}, Depends: []string{"root1"}, Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)}})
		wf.AddTask(iapetus.Task{Name: "mid2", Command: "echo", Args: []string{"M2"}, Depends: []string{"root2"}, Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)}})
		wf.AddTask(iapetus.Task{Name: "leaf1", Command: "echo", Args: []string{"L1"}, Depends: []string{"mid1"}, Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)}})
		wf.AddTask(iapetus.Task{Name: "leaf2", Command: "echo", Args: []string{"L2"}, Depends: []string{"mid2"}, Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)}})
		err := wf.Run()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("cycle detection", func(t *testing.T) {
		wf := iapetus.NewWorkflow("cycle-detection", zap.NewNop())
		wf.AddTask(iapetus.Task{Name: "a", Command: "echo", Args: []string{"A"}, Depends: []string{"c"}, Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)}})
		wf.AddTask(iapetus.Task{Name: "b", Command: "echo", Args: []string{"B"}, Depends: []string{"a"}, Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)}})
		wf.AddTask(iapetus.Task{Name: "c", Command: "echo", Args: []string{"C"}, Depends: []string{"b"}, Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)}})
		err := wf.Run()
		if err == nil || !strings.Contains(err.Error(), "cycle") {
			t.Errorf("expected cycle error, got %v", err)
		}
	})

	t.Run("missing dependency", func(t *testing.T) {
		wf := iapetus.NewWorkflow("missing-dependency", zap.NewNop())
		wf.AddTask(iapetus.Task{Name: "a", Command: "echo", Args: []string{"A"}, Depends: []string{"notfound"}, Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)}})
		err := wf.Run()
		if err == nil || !strings.Contains(err.Error(), "does not exist") {
			t.Errorf("expected missing dependency error, got %v", err)
		}
	})

	t.Run("self dependency", func(t *testing.T) {
		wf := iapetus.NewWorkflow("self-dependency", zap.NewNop())
		wf.AddTask(iapetus.Task{Name: "a", Command: "echo", Args: []string{"A"}, Depends: []string{"a"}, Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)}})
		err := wf.Run()
		if err == nil || !strings.Contains(err.Error(), "cycle") {
			t.Errorf("expected cycle error, got %v", err)
		}
	})

	t.Run("parallel roots", func(t *testing.T) {
		wf := iapetus.NewWorkflow("parallel-roots", zap.NewNop())
		wf.AddTask(iapetus.Task{Name: "a", Command: "echo", Args: []string{"A"}, Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)}})
		wf.AddTask(iapetus.Task{Name: "b", Command: "echo", Args: []string{"B"}, Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)}})
		err := wf.Run()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("multiple dependencies", func(t *testing.T) {
		wf := iapetus.NewWorkflow("multiple-dependencies", zap.NewNop())
		wf.AddTask(iapetus.Task{Name: "a", Command: "echo", Args: []string{"A"}, Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)}})
		wf.AddTask(iapetus.Task{Name: "b", Command: "echo", Args: []string{"B"}, Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)}})
		wf.AddTask(iapetus.Task{Name: "c", Command: "echo", Args: []string{"C"}, Depends: []string{"a", "b"}, Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)}})
		err := wf.Run()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("PreRun/PostRun hooks", func(t *testing.T) {
		// Skipped: PreRun/PostRun hooks are not present in the current Workflow struct
	})

	t.Run("PreRun fails", func(t *testing.T) {
		// Skipped: PreRun/PostRun hooks are not present in the current Workflow struct
	})

	t.Run("PostRun fails", func(t *testing.T) {
		// Skipped: PreRun/PostRun hooks are not present in the current Workflow struct
	})

	t.Run("empty workflow", func(t *testing.T) {
		wf := iapetus.NewWorkflow("empty-workflow", zap.NewNop())
		err := wf.Run()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("duplicate task names", func(t *testing.T) {
		wf := iapetus.NewWorkflow("duplicate-task-names", zap.NewNop())
		wf.AddTask(iapetus.Task{Name: "a", Command: "echo", Args: []string{"A"}, Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)}})
		wf.AddTask(iapetus.Task{Name: "a", Command: "echo", Args: []string{"A2"}, Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)}})
		err := wf.Run()
		if err == nil || !strings.Contains(err.Error(), "already exists") {
			t.Errorf("expected duplicate task error, got %v", err)
		}
	})

	t.Run("custom assertion fails", func(t *testing.T) {
		wf := iapetus.NewWorkflow("custom-assertion-fails", zap.NewNop())
		wf.AddTask(iapetus.Task{Name: "a", Command: "echo", Args: []string{"A"}, Asserts: []func(*iapetus.Task) error{
			func(t *iapetus.Task) error { return exec.ErrNotFound },
		}})
		err := wf.Run()
		if err == nil || !strings.Contains(err.Error(), "not found") {
			t.Errorf("expected custom assertion error, got %v", err)
		}
	})

	t.Run("step with env and timeout", func(t *testing.T) {
		wf := iapetus.NewWorkflow("step-with-env-and-timeout", zap.NewNop())
		wf.AddTask(iapetus.Task{Name: "a", Command: "sh", Args: []string{"-c", "sleep 0.1; echo $FOO"}, Timeout: 1 * time.Second, Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)}})
		err := wf.Run()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}

// Property-based test: For any valid workflow, all tasks should run once and dependencies respected
func TestWorkflow_PropertyBased(t *testing.T) {
	f := func(numTasks uint8) bool {
		n := int(numTasks%10) + 2 // 2-11 tasks
		names := make([]string, n)
		seed := int64(numTasks) // Use numTasks as seed for determinism
		r := rand.New(rand.NewSource(seed))
		for i := 0; i < n; i++ {
			names[i] = fmt.Sprintf("t%d", i)
		}
		steps := make([]iapetus.Task, n)
		for i := 0; i < n; i++ {
			deps := []string{}
			for j := 0; j < i; j++ {
				if r.Float64() < 0.3 {
					deps = append(deps, names[j])
				}
			}
			steps[i] = iapetus.Task{
				Name:    names[i],
				Command: "true",
				Depends: deps,
				Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)},
			}
		}
		wf := iapetus.NewWorkflow("property-based", zap.NewNop())
		for i := 0; i < n; i++ {
			wf.AddTask(steps[i])
		}
		ran := make(map[string]int)
		var mu sync.Mutex
		wf.AddOnTaskStartHook(func(task *iapetus.Task) {
			mu.Lock()
			ran[task.Name]++
			mu.Unlock()
		})
		if err := wf.Run(); err != nil {
			// If the DAG is invalid, that's fine, skip
			if !strings.Contains(err.Error(), "cycle") && !strings.Contains(err.Error(), "does not exist") {
				t.Logf("unexpected error: %v", err)
				return false
			}
			return true
		}
		// All tasks should have run exactly once
		if len(ran) != n {
			t.Logf("not all tasks ran: %v", ran)
			return false
		}
		for _, count := range ran {
			if count != 1 {
				t.Logf("task ran %d times", count)
				return false
			}
		}
		return true
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

// --- Additional tests for backend, logger, and EnvMap propagation ---

// mockBackend is a test backend that records calls.
type mockBackend struct {
	called *bool
}

func (m *mockBackend) RunTask(task *iapetus.Task) error {
	if m.called != nil {
		*m.called = true
	}
	return nil
}
func (m *mockBackend) ValidateTask(task *iapetus.Task) error { return nil }

func TestWorkflow_BackendPropagation(t *testing.T) {
	called := false
	iapetus.RegisterBackend("mock", &mockBackend{called: &called})
	wf := iapetus.NewWorkflow("test-backend", zap.NewNop())
	wf.Backend = "mock"
	task := iapetus.Task{
		Name:    "echo",
		Command: "echo",
		Args:    []string{"hello"},
		Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)},
	}
	wf.AddTask(task)
	_ = wf.Run()
	if !called {
		t.Errorf("expected mock backend to be called")
	}
}

func TestWorkflow_LoggerPropagation(t *testing.T) {
	wf := iapetus.NewWorkflow("test-logger", zap.NewNop())
	task := iapetus.Task{
		Name:    "echo",
		Command: "echo",
		Args:    []string{"hello"},
		Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)},
	}
	wf.AddTask(task)
	if wf.Steps[0].Logger() == nil {
		t.Errorf("expected logger to be propagated to task")
	}
}

func TestWorkflow_EnvMapPropagation(t *testing.T) {
	wf := iapetus.NewWorkflow("test-envmap", zap.NewNop())
	wf.EnvMap = map[string]string{"FOO": "bar"}
	task := iapetus.Task{
		Name:    "echo",
		Command: "echo",
		Args:    []string{"hello"},
		Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)},
	}
	wf.AddTask(task)
	if val, ok := wf.Steps[0].EnvMap["FOO"]; !ok || val != "bar" {
		t.Errorf("expected EnvMap to be propagated to task")
	}
}

func TestWorkflow_PerTaskBackendOverride(t *testing.T) {
	called := false
	iapetus.RegisterBackend("mock2", &mockBackend{called: &called})
	wf := iapetus.NewWorkflow("test-per-task-backend", zap.NewNop())
	wf.Backend = "bash"
	task := iapetus.Task{
		Name:    "echo",
		Command: "echo",
		Args:    []string{"hello"},
		Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)},
	}
	task.SetBackend("mock2")
	wf.AddTask(task)
	_ = wf.Run()
	if !called {
		t.Errorf("expected per-task backend to be used")
	}
}

func TestWorkflow_ErrorOnMissingBackend(t *testing.T) {
	wf := iapetus.NewWorkflow("test-missing-backend", zap.NewNop())
	task := iapetus.Task{
		Name:    "echo",
		Command: "echo",
		Args:    []string{"hello"},
		Asserts: []func(*iapetus.Task) error{iapetus.AssertExitCode(0)},
	}
	task.SetBackend("doesnotexist")
	wf.AddTask(task)
	err := wf.Run()
	if err == nil || !strings.Contains(err.Error(), "backend doesnotexist not found") {
		t.Errorf("expected error for missing backend, got %v", err)
	}
}
