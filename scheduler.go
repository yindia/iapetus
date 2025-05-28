package iapetus

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// schedulerEvent represents an event in the scheduler loop.
type schedulerEvent struct {
	eventType string // e.g. "ready", "done", "cancel"
	name      string // task name, if relevant
}

// dagScheduler encapsulates all state for parallel DAG execution.
type dagScheduler struct {
	w          *Workflow
	taskMap    map[string]*Task
	depCount   map[string]int
	dependents map[string][]string
	wg         sync.WaitGroup
	mu         sync.Mutex
	errOnce    error
	ctx        context.Context
	cancel     context.CancelFunc
	doneCh     chan struct{}
	completed  map[string]bool
	started    map[string]bool
	cancelled  bool
	eventCh    chan schedulerEvent
}

// newDagScheduler initializes the scheduler state from the task order.
func newDagScheduler(w *Workflow, order []*Task) *dagScheduler {
	taskMap := make(map[string]*Task)
	depCount := make(map[string]int)
	dependents := make(map[string][]string)
	for _, t := range order {
		taskMap[t.Name] = t
		depCount[t.Name] = len(t.Depends)
		for _, dep := range t.Depends {
			dependents[dep] = append(dependents[dep], t.Name)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &dagScheduler{
		w:          w,
		taskMap:    taskMap,
		depCount:   depCount,
		dependents: dependents,
		errOnce:    nil,
		ctx:        ctx,
		cancel:     cancel,
		doneCh:     make(chan struct{}, 1),
		completed:  make(map[string]bool),
		started:    make(map[string]bool),
		cancelled:  false,
		eventCh:    make(chan schedulerEvent, len(order)*2),
	}
}

// run executes the DAG in parallel, respecting dependencies.
func (s *dagScheduler) run() error {
	defer s.cancel()
	// Seed event queue with tasks that have no dependencies
	for name, count := range s.depCount {
		if count == 0 {
			s.eventCh <- schedulerEvent{eventType: "ready", name: name}
		}
	}
	if len(s.taskMap) == 0 {
		s.w.logger.Debug("Scheduler: no tasks to run, exiting immediately", zap.String("workflow", s.w.Name))
		return nil
	}

	for {
		select {
		case <-s.ctx.Done():
			s.cancelled = true
			s.w.logger.Debug("Scheduler: context cancelled, breaking main loop", zap.String("workflow", s.w.Name))
			return s.errOnce

		case <-s.doneCh:
			s.w.logger.Debug("Scheduler: doneCh signaled, breaking main loop", zap.String("workflow", s.w.Name))
			return s.errOnce

		case ev := <-s.eventCh:
			switch ev.eventType {
			case "ready":
				s.handleReady(ev.name)
			case "done":
				s.w.logger.Debug("All tasks completed, breaking scheduler loop", zap.String("workflow", s.w.Name))
				return s.errOnce
			case "cancel":
				s.cancelled = true
				s.w.logger.Debug("Scheduler: cancel event, breaking main loop", zap.String("workflow", s.w.Name))
				return s.errOnce
			}
		case <-time.After(10 * time.Millisecond):
			s.mu.Lock()
			allCompleted := len(s.completed) == len(s.taskMap)
			s.mu.Unlock()
			if allCompleted {
				s.eventCh <- schedulerEvent{eventType: "done"}
			}
		}
	}
}

// handleReady handles a ready event by starting the task if not already started.
func (s *dagScheduler) handleReady(name string) {
	s.mu.Lock()
	task, ok := s.taskMap[name]
	if !ok || s.started[name] {
		s.mu.Unlock()
		return
	}
	s.started[name] = true
	s.mu.Unlock()
	s.wg.Add(1)
	go s.runTask(name, task)
}

// runTask executes a single task and handles completion, dependents, and error propagation.
func (s *dagScheduler) runTask(name string, task *Task) {
	defer s.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("panic in task %s: %v", name, r)
			s.mu.Lock()
			s.w.OnTaskFailure(task, err)
			if s.errOnce == nil {
				s.errOnce = &WorkflowError{
					StepName:     task.Name,
					WorkflowName: s.w.Name,
					Err:          err,
				}
				s.cancel()
			}
			s.completed[name] = true
			if len(s.completed) == len(s.taskMap) {
				select {
				case s.doneCh <- struct{}{}:
				default:
				}
			}
			s.mu.Unlock()
			s.w.OnTaskComplete(task)
			s.w.logger.Debug("Task completed (panic)", zap.String("task", task.Name))
		}
	}()
	s.w.OnTaskStart(task)
	err := task.Run()
	s.mu.Lock()
	if err != nil {
		s.w.OnTaskFailure(task, err)
		if s.errOnce == nil {
			s.errOnce = &WorkflowError{
				StepName:     task.Name,
				WorkflowName: s.w.Name,
				Err:          err,
			}
			s.cancel()
		}
	} else {
		s.w.OnTaskSuccess(task)
	}
	s.completed[name] = true
	if len(s.completed) == len(s.taskMap) {
		select {
		case s.doneCh <- struct{}{}:
		default:
		}
	}
	s.mu.Unlock()
	s.w.OnTaskComplete(task)
	s.w.logger.Debug("Task completed", zap.String("task", task.Name))

	s.mu.Lock()
	for _, dep := range s.dependents[name] {
		s.depCount[dep]--
		if s.depCount[dep] == 0 && !s.started[dep] {
			s.eventCh <- schedulerEvent{eventType: "ready", name: dep}
		}
	}
	s.mu.Unlock()
}
