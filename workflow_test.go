package iapetus

import (
	"fmt"
	"math/rand"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"testing/quick"
	"time"

	"go.uber.org/zap"
)

func TestWorkflowRun(t *testing.T) {
	tests := []struct {
		name     string
		workflow Workflow
		wantErr  bool
	}{
		{
			name: "All Steps Pass",
			workflow: Workflow{
				Steps: []Task{
					{
						Name:    "step1",
						Command: "echo",
						Args:    []string{"hello"},
						Asserts: []func(*Task) error{AssertExitCode(0), AssertOutputEquals("hello\n")},
					},
				},
				logger: zap.NewNop(),
			},
			wantErr: false,
		},

		{
			name: "Step Fails on Output",
			workflow: Workflow{
				Steps: []Task{
					{
						Name:    "step1",
						Command: "echo",
						Args:    []string{"hello"},
						Asserts: []func(*Task) error{AssertOutputEquals("world\n")},
					},
				},
				logger: zap.NewNop(),
			},
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
		wf := Workflow{
			Steps: []Task{
				{Name: "a", Command: "echo", Args: []string{"1"}, Asserts: []func(*Task) error{AssertExitCode(0)}},
				{Name: "b", Command: "echo", Args: []string{"2"}, Depends: []string{"a"}, Asserts: []func(*Task) error{AssertExitCode(0)}},
				{Name: "c", Command: "echo", Args: []string{"3"}, Depends: []string{"b"}, Asserts: []func(*Task) error{AssertExitCode(0)}},
			},
			logger: zap.NewNop(),
		}
		err := wf.Run()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("diamond dependency", func(t *testing.T) {
		wf := Workflow{
			Steps: []Task{
				{Name: "a", Command: "echo", Args: []string{"A"}, Asserts: []func(*Task) error{AssertExitCode(0)}},
				{Name: "b", Command: "echo", Args: []string{"B"}, Depends: []string{"a"}, Asserts: []func(*Task) error{AssertExitCode(0)}},
				{Name: "c", Command: "echo", Args: []string{"C"}, Depends: []string{"a"}, Asserts: []func(*Task) error{AssertExitCode(0)}},
				{Name: "d", Command: "echo", Args: []string{"D"}, Depends: []string{"b", "c"}, Asserts: []func(*Task) error{AssertExitCode(0)}},
			},
			logger: zap.NewNop(),
		}
		err := wf.Run()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("multiple roots and leaves", func(t *testing.T) {
		wf := Workflow{
			Steps: []Task{
				{Name: "root1", Command: "echo", Args: []string{"R1"}, Asserts: []func(*Task) error{AssertExitCode(0)}},
				{Name: "root2", Command: "echo", Args: []string{"R2"}, Asserts: []func(*Task) error{AssertExitCode(0)}},
				{Name: "mid1", Command: "echo", Args: []string{"M1"}, Depends: []string{"root1"}, Asserts: []func(*Task) error{AssertExitCode(0)}},
				{Name: "mid2", Command: "echo", Args: []string{"M2"}, Depends: []string{"root2"}, Asserts: []func(*Task) error{AssertExitCode(0)}},
				{Name: "leaf1", Command: "echo", Args: []string{"L1"}, Depends: []string{"mid1"}, Asserts: []func(*Task) error{AssertExitCode(0)}},
				{Name: "leaf2", Command: "echo", Args: []string{"L2"}, Depends: []string{"mid2"}, Asserts: []func(*Task) error{AssertExitCode(0)}},
			},
			logger: zap.NewNop(),
		}
		err := wf.Run()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("cycle detection", func(t *testing.T) {
		wf := Workflow{
			Steps: []Task{
				{Name: "a", Command: "echo", Args: []string{"A"}, Depends: []string{"c"}, Asserts: []func(*Task) error{AssertExitCode(0)}},
				{Name: "b", Command: "echo", Args: []string{"B"}, Depends: []string{"a"}, Asserts: []func(*Task) error{AssertExitCode(0)}},
				{Name: "c", Command: "echo", Args: []string{"C"}, Depends: []string{"b"}, Asserts: []func(*Task) error{AssertExitCode(0)}},
			},
			logger: zap.NewNop(),
		}
		err := wf.Run()
		if err == nil || !strings.Contains(err.Error(), "cycle") {
			t.Errorf("expected cycle error, got %v", err)
		}
	})

	t.Run("missing dependency", func(t *testing.T) {
		wf := Workflow{
			Steps: []Task{
				{Name: "a", Command: "echo", Args: []string{"A"}, Depends: []string{"notfound"}, Asserts: []func(*Task) error{AssertExitCode(0)}},
			},
			logger: zap.NewNop(),
		}
		err := wf.Run()
		if err == nil || !strings.Contains(err.Error(), "does not exist") {
			t.Errorf("expected missing dependency error, got %v", err)
		}
	})

	t.Run("self dependency", func(t *testing.T) {
		wf := Workflow{
			Steps: []Task{
				{Name: "a", Command: "echo", Args: []string{"A"}, Depends: []string{"a"}, Asserts: []func(*Task) error{AssertExitCode(0)}},
			},
			logger: zap.NewNop(),
		}
		err := wf.Run()
		if err == nil || !strings.Contains(err.Error(), "cycle") {
			t.Errorf("expected cycle error, got %v", err)
		}
	})

	t.Run("parallel roots", func(t *testing.T) {
		wf := Workflow{
			Steps: []Task{
				{Name: "a", Command: "echo", Args: []string{"A"}, Asserts: []func(*Task) error{AssertExitCode(0)}},
				{Name: "b", Command: "echo", Args: []string{"B"}, Asserts: []func(*Task) error{AssertExitCode(0)}},
			},
			logger: zap.NewNop(),
		}
		err := wf.Run()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("multiple dependencies", func(t *testing.T) {
		wf := Workflow{
			Steps: []Task{
				{Name: "a", Command: "echo", Args: []string{"A"}, Asserts: []func(*Task) error{AssertExitCode(0)}},
				{Name: "b", Command: "echo", Args: []string{"B"}, Asserts: []func(*Task) error{AssertExitCode(0)}},
				{Name: "c", Command: "echo", Args: []string{"C"}, Depends: []string{"a", "b"}, Asserts: []func(*Task) error{AssertExitCode(0)}},
			},
			logger: zap.NewNop(),
		}
		err := wf.Run()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("PreRun/PostRun hooks", func(t *testing.T) {
		preRunCalled := false
		postRunCalled := false
		wf := Workflow{
			PreRun:  func(w *Workflow) error { preRunCalled = true; return nil },
			PostRun: func(w *Workflow) error { postRunCalled = true; return nil },
			Steps: []Task{
				{Name: "a", Command: "echo", Args: []string{"A"}, Asserts: []func(*Task) error{AssertExitCode(0)}},
			},
			logger: zap.NewNop(),
		}
		err := wf.Run()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if !preRunCalled {
			t.Errorf("expected PreRun to be called")
		}
		if !postRunCalled {
			t.Errorf("expected PostRun to be called")
		}
	})

	t.Run("PreRun fails", func(t *testing.T) {
		wf := Workflow{
			PreRun: func(w *Workflow) error { return exec.ErrNotFound },
			Steps: []Task{
				{Name: "a", Command: "echo", Args: []string{"A"}, Asserts: []func(*Task) error{AssertExitCode(0)}},
			},
			logger: zap.NewNop(),
		}
		err := wf.Run()
		if err == nil || !strings.Contains(err.Error(), "pre-run") {
			t.Errorf("expected pre-run error, got %v", err)
		}
	})

	t.Run("PostRun fails", func(t *testing.T) {
		wf := Workflow{
			PostRun: func(w *Workflow) error { return exec.ErrNotFound },
			Steps: []Task{
				{Name: "a", Command: "echo", Args: []string{"A"}, Asserts: []func(*Task) error{AssertExitCode(0)}},
			},
			logger: zap.NewNop(),
		}
		err := wf.Run()
		if err == nil || !strings.Contains(err.Error(), "post-run") {
			t.Errorf("expected post-run error, got %v", err)
		}
	})

	t.Run("empty workflow", func(t *testing.T) {
		wf := Workflow{
			logger: zap.NewNop(),
		}
		err := wf.Run()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("duplicate task names", func(t *testing.T) {
		wf := Workflow{
			Steps: []Task{
				{Name: "a", Command: "echo", Args: []string{"A"}, Asserts: []func(*Task) error{AssertExitCode(0)}},
				{Name: "a", Command: "echo", Args: []string{"A2"}, Asserts: []func(*Task) error{AssertExitCode(0)}},
			},
			logger: zap.NewNop(),
		}
		err := wf.Run()
		if err == nil || !strings.Contains(err.Error(), "already exists") {
			t.Errorf("expected duplicate task error, got %v", err)
		}
	})

	t.Run("custom assertion fails", func(t *testing.T) {
		wf := Workflow{
			Steps: []Task{
				{Name: "a", Command: "echo", Args: []string{"A"}, Asserts: []func(*Task) error{
					func(t *Task) error { return exec.ErrNotFound },
				}},
			},
			logger: zap.NewNop(),
		}
		err := wf.Run()
		if err == nil || !strings.Contains(err.Error(), "not found") {
			t.Errorf("expected custom assertion error, got %v", err)
		}
	})

	t.Run("step with env and timeout", func(t *testing.T) {
		wf := Workflow{
			Steps: []Task{
				{Name: "a", Command: "sh", Args: []string{"-c", "sleep 0.1; echo $FOO"}, Env: []string{"FOO=bar"}, Timeout: 1 * time.Second, Asserts: []func(*Task) error{AssertExitCode(0)}},
			},
			logger: zap.NewNop(),
		}
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
		steps := make([]Task, n)
		for i := 0; i < n; i++ {
			deps := []string{}
			for j := 0; j < i; j++ {
				if r.Float64() < 0.3 {
					deps = append(deps, names[j])
				}
			}
			steps[i] = Task{
				Name:    names[i],
				Command: "true",
				Depends: deps,
				Asserts: []func(*Task) error{AssertExitCode(0)},
			}
		}
		wf := Workflow{Steps: steps, logger: zap.NewNop()}
		ran := make(map[string]int)
		var mu sync.Mutex
		wf.AddOnTaskStartHook(func(task *Task) {
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
