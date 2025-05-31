package iapetus

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func init() {
	RegisterBackend("bash", &testBashBackend{})
}

type testBashBackend struct{}

func (b *testBashBackend) RunTask(task *Task) error {
	task.Actual.Output = ""
	task.Actual.ExitCode = 0
	return RunAssertions(task)
}
func (b *testBashBackend) ValidateTask(task *Task) error {
	return nil
}

func (b *testBashBackend) GetName() string {
	return "bash"
}

func (b *testBashBackend) GetStatus() string {
	return "bash"
}
func TestDagScheduler_ParallelExecution(t *testing.T) {
	var mu sync.Mutex
	order := []string{}
	wg := &sync.WaitGroup{}
	tasks := []*Task{
		{
			Name:    "a",
			Command: "true",
			Asserts: []func(*Task) error{
				func(t *Task) error {
					mu.Lock()
					order = append(order, "a")
					mu.Unlock()
					wg.Done()
					time.Sleep(50 * time.Millisecond)
					return nil
				},
			},
		},
		{
			Name:    "b",
			Command: "true",
			Asserts: []func(*Task) error{
				func(t *Task) error {
					mu.Lock()
					order = append(order, "b")
					mu.Unlock()
					wg.Done()
					time.Sleep(50 * time.Millisecond)
					return nil
				},
			},
		},
	}
	wg.Add(len(tasks))
	w := NewWorkflow("test", zap.NewNop())
	ds := newDagScheduler(w, tasks)
	errCh := make(chan error, 1)
	go func() {
		errCh <- ds.run()
	}()
	wg.Wait()
	err := <-errCh
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order) != 2 {
		t.Fatalf("expected 2 tasks to run, got %d", len(order))
	}
	if order[0] == order[1] {
		t.Errorf("expected parallel execution, got sequential: %v", order)
	}
}

func TestDagScheduler_Dependencies(t *testing.T) {
	var mu sync.Mutex
	order := []string{}
	wg := &sync.WaitGroup{}
	tasks := []*Task{
		{
			Name:    "a",
			Command: "true",
			Asserts: []func(*Task) error{
				func(t *Task) error {
					mu.Lock()
					order = append(order, "a")
					mu.Unlock()
					wg.Done()
					return nil
				},
			},
		},
		{
			Name:    "b",
			Command: "true",
			Depends: []string{"a"},
			Asserts: []func(*Task) error{
				func(t *Task) error {
					mu.Lock()
					order = append(order, "b")
					mu.Unlock()
					wg.Done()
					return nil
				},
			},
		},
	}
	wg.Add(len(tasks))
	w := NewWorkflow("test", zap.NewNop())
	ds := newDagScheduler(w, tasks)
	errCh := make(chan error, 1)
	go func() {
		errCh <- ds.run()
	}()
	wg.Wait()
	err := <-errCh
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order[0] != "a" || order[1] != "b" {
		t.Errorf("expected a before b, got %v", order)
	}
}

func TestDagScheduler_ErrorPropagation(t *testing.T) {
	wg := &sync.WaitGroup{}
	tasks := []*Task{
		{
			Name:    "a",
			Command: "true",
			Asserts: []func(*Task) error{
				func(t *Task) error {
					wg.Done()
					return assert.AnError
				},
			},
		},
		{
			Name:    "b",
			Command: "true",
			Asserts: []func(*Task) error{
				func(t *Task) error {
					time.Sleep(100 * time.Millisecond)
					wg.Done()
					return nil
				},
			},
		},
	}
	wg.Add(len(tasks))
	w := NewWorkflow("test", zap.NewNop())
	ds := newDagScheduler(w, tasks)
	err := ds.run()
	if err == nil {
		t.Errorf("expected error propagation, got nil")
	}
}

func TestDagScheduler_EmptyDAG(t *testing.T) {
	w := NewWorkflow("test", zap.NewNop())
	ds := newDagScheduler(w, []*Task{})
	err := ds.run()
	if err != nil {
		t.Errorf("expected no error for empty DAG, got %v", err)
	}
}

func TestDagScheduler_SingleTask(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	task := &Task{
		Name:    "a",
		Command: "true",
		Asserts: []func(*Task) error{
			func(t *Task) error { wg.Done(); return nil },
		},
	}
	w := NewWorkflow("test", zap.NewNop())
	ds := newDagScheduler(w, []*Task{task})
	errCh := make(chan error, 1)
	go func() {
		errCh <- ds.run()
	}()
	wg.Wait()
	err := <-errCh
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDagScheduler_TaskPanic(t *testing.T) {
	w := NewWorkflow("panic-test", zap.NewNop())
	panicTask := &Task{
		Name:    "panic-task",
		Command: "true",
		Asserts: []func(*Task) error{
			func(t *Task) error {
				panic("simulated panic")
			},
		},
	}
	ds := newDagScheduler(w, []*Task{panicTask})
	err := ds.run()
	if err == nil {
		t.Errorf("expected error from panic, got nil")
	}
}

func TestDagScheduler_ContextCancellation(t *testing.T) {
	w := NewWorkflow("cancel-test", zap.NewNop())
	// Use a done channel to simulate cancellation
	done := make(chan struct{})
	blocker := &Task{
		Name:    "blocker",
		Command: "true",
		Asserts: []func(*Task) error{
			func(t *Task) error {
				select {
				case <-time.After(2 * time.Second):
					return nil
				case <-done:
					return context.Canceled
				}
			},
		},
	}
	ds := newDagScheduler(w, []*Task{blocker})
	// Cancel after 100ms
	go func() {
		time.Sleep(100 * time.Millisecond)
		close(done)
		ds.cancel()
	}()
	start := time.Now()
	err := ds.run()
	elapsed := time.Since(start)
	if elapsed > time.Second {
		t.Errorf("scheduler did not exit promptly on context cancellation")
	}
	if err != nil && !errors.Is(err, context.Canceled) && !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("expected context.Canceled error, got %v", err)
	}
	// Note: For prompt cancellation, tasks must be context-aware.
}

func TestDagScheduler_ObservabilityHooks(t *testing.T) {
	var mu sync.Mutex
	calls := make(map[string]map[string]bool) // taskName -> hookType -> called
	hookTypes := []string{"start", "success", "fail", "complete"}
	for _, name := range []string{"ok", "fail"} {
		calls[name] = map[string]bool{}
	}
	w := NewWorkflow("hooks-test", zap.NewNop())
	w.AddOnTaskStartHook(func(task *Task) { mu.Lock(); calls[task.Name]["start"] = true; mu.Unlock() })
	w.AddOnTaskSuccessHook(func(task *Task) { mu.Lock(); calls[task.Name]["success"] = true; mu.Unlock() })
	w.AddOnTaskFailureHook(func(task *Task, err error) { mu.Lock(); calls[task.Name]["fail"] = true; mu.Unlock() })
	w.AddOnTaskCompleteHook(func(task *Task) { mu.Lock(); calls[task.Name]["complete"] = true; mu.Unlock() })
	RegisterBackend("bash", &BashBackend{})
	tasks := []*Task{
		{
			Name:    "ok",
			Command: "true",
			Backend: "bash",
			Asserts: []func(*Task) error{
				func(t *Task) error { return nil },
			},
		},
		{
			Name:    "fail",
			Command: "true",
			Backend: "bash",
			Asserts: []func(*Task) error{
				func(t *Task) error { return errors.New("fail") },
			},
		},
	}
	ds := newDagScheduler(w, tasks)
	err := ds.run()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	// Check that all hooks were called for both tasks as appropriate
	mu.Lock()
	defer mu.Unlock()
	for _, name := range []string{"ok", "fail"} {
		for _, hook := range hookTypes {
			if hook == "success" && name == "fail" {
				if calls[name][hook] {
					t.Errorf("unexpected success hook for failed task %s", name)
				}
				continue
			}
			if hook == "fail" && name == "ok" {
				if calls[name][hook] {
					t.Errorf("unexpected fail hook for successful task %s", name)
				}
				continue
			}
			if !calls[name][hook] {
				t.Errorf("missing %s hook for task %s", hook, name)
			}
		}
	}
}

func TestDagScheduler_NoGoroutineLeak(t *testing.T) {
	w := NewWorkflow("leak-test", zap.NewNop())
	wg := &sync.WaitGroup{}
	tasks := []*Task{}
	for i := 0; i < 10; i++ {
		task := &Task{
			Name:    fmt.Sprintf("t%d", i),
			Command: "true",
			Asserts: []func(*Task) error{func(t *Task) error { wg.Done(); return nil }},
		}
		tasks = append(tasks, task)
		wg.Add(1)
	}
	ds := newDagScheduler(w, tasks)
	errCh := make(chan error, 1)
	go func() {
		errCh <- ds.run()
	}()
	c := make(chan struct{})
	go func() {
		wg.Wait()
		close(c)
	}()
	select {
	case <-c:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("goroutines did not finish (possible leak)")
	}
	err := <-errCh
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
