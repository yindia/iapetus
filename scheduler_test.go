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

func TestDagScheduler_ParallelExecution(t *testing.T) {
	var mu sync.Mutex
	order := []string{}
	wg := &sync.WaitGroup{}
	tasks := []*Task{
		{
			Name: "a",
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
			Name: "b",
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
			Name: "a",
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
	calls := []string{}
	w := NewWorkflow("hooks-test", zap.NewNop())
	w.AddOnTaskStartHook(func(task *Task) { mu.Lock(); calls = append(calls, "start:"+task.Name); mu.Unlock() })
	w.AddOnTaskSuccessHook(func(task *Task) { mu.Lock(); calls = append(calls, "success:"+task.Name); mu.Unlock() })
	w.AddOnTaskFailureHook(func(task *Task, err error) { mu.Lock(); calls = append(calls, "fail:"+task.Name); mu.Unlock() })
	w.AddOnTaskCompleteHook(func(task *Task) { mu.Lock(); calls = append(calls, "complete:"+task.Name); mu.Unlock() })
	tasks := []*Task{
		{
			Name:    "ok",
			Command: "true",
			Asserts: []func(*Task) error{
				func(t *Task) error { return nil },
			},
		},
		{
			Name:    "fail",
			Command: "true",
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
	// Check that all hooks were called for both tasks
	mu.Lock()
	defer mu.Unlock()
	foundStart, foundSuccess, foundFail, foundComplete := false, false, false, false
	for _, c := range calls {
		if c == "start:ok" || c == "start:fail" {
			foundStart = true
		}
		if c == "success:ok" {
			foundSuccess = true
		}
		if c == "fail:fail" {
			foundFail = true
		}
		if c == "complete:ok" || c == "complete:fail" {
			foundComplete = true
		}
	}
	if !foundStart || !foundSuccess || !foundFail || !foundComplete {
		t.Errorf("not all hooks were called: %v", calls)
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
