package iapetus

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
)

// Simple stress test: many workflows, many tasks, all succeed
func TestWorkflow_SimpleStress(t *testing.T) {
	t.Parallel()
	const numWorkflows = 100
	const tasksPerWorkflow = 20
	var wg sync.WaitGroup
	for i := 0; i < numWorkflows; i++ {
		wg.Add(1)
		go func(wfID int) {
			defer wg.Done()
			steps := make([]Task, 0, tasksPerWorkflow)
			for j := 0; j < tasksPerWorkflow; j++ {
				name := fmt.Sprintf("w%d-t%d", wfID, j)
				depends := []string{}
				if j > 0 {
					depends = []string{fmt.Sprintf("w%d-t%d", wfID, j-1)}
				}
				steps = append(steps, Task{
					Name:     name,
					Command:  "sh",
					Args:     []string{"-c", "sleep 0.001"},
					Depends:  depends,
					Expected: Output{ExitCode: 0},
					Asserts:  []func(*Task) error{AssertByExitCode},
				})
			}
			wf := Workflow{Steps: steps, logger: zap.NewNop()}
			if err := wf.Run(); err != nil {
				t.Errorf("workflow %d failed: %v", wfID, err)
			}
		}(i)
	}
	wg.Wait()
}

// Advanced stress test: random DAGs, random failures, panics
func TestWorkflow_AdvancedStress(t *testing.T) {
	t.Parallel()
	seed := time.Now().UnixNano()
	r := rand.New(rand.NewSource(seed))
	const numWorkflows = 20
	const tasksPerWorkflow = 50
	var wg sync.WaitGroup
	for i := 0; i < numWorkflows; i++ {
		wg.Add(1)
		go func(wfID int) {
			defer wg.Done()
			steps := make([]Task, 0, tasksPerWorkflow)
			for j := 0; j < tasksPerWorkflow; j++ {
				name := fmt.Sprintf("w%d-t%d", wfID, j)
				// Random dependencies: up to 3 from previous tasks
				depends := []string{}
				for d := 0; d < r.Intn(3); d++ {
					if k := r.Intn(j + 1); k < j {
						depends = append(depends, fmt.Sprintf("w%d-t%d", wfID, k))
					}
				}
				// Randomly fail or panic
				asserts := []func(*Task) error{AssertByExitCode}
				if r.Float64() < 0.05 {
					asserts = append(asserts, func(t *Task) error { return fmt.Errorf("random fail") })
				}
				if r.Float64() < 0.02 {
					asserts = append(asserts, func(t *Task) error { panic("random panic") })
				}
				steps = append(steps, Task{
					Name:     name,
					Command:  "sh",
					Args:     []string{"-c", "sleep 0.001"},
					Depends:  depends,
					Expected: Output{ExitCode: 0},
					Asserts:  asserts,
				})
			}
			wf := Workflow{Steps: steps, logger: zap.NewNop()}
			_ = wf.Run() // Ignore error: we expect some failures
		}(i)
	}
	wg.Wait()
}
