package iapetus

import (
	"fmt"
	"sync"
)

// TaskNode wraps a Task and tracks dependencies and dependents.
type TaskNode struct {
	Name    string
	Task    *Task
	Deps    []string
	Depends []string
}

// DAG represents a thread-safe directed acyclic graph of tasks.
type DAG struct {
	nodes map[string]*TaskNode
	edges map[string][]string
	mu    sync.RWMutex
}

// NewDag creates a new, empty DAG.
func NewDag() *DAG {
	return &DAG{
		nodes: make(map[string]*TaskNode),
		edges: make(map[string][]string),
	}
}

// AddTask adds a task to the DAG. Allows adding in any order. No missing dependency check here.
func (d *DAG) AddTask(task *Task) error {
	if task == nil {
		return fmt.Errorf("task is nil")
	}
	if task.Name == "" {
		return fmt.Errorf("task must have a name")
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, exists := d.nodes[task.Name]; exists {
		return fmt.Errorf("task with name %s already exists", task.Name)
	}
	d.nodes[task.Name] = &TaskNode{
		Name:    task.Name,
		Task:    task,
		Deps:    append([]string{}, task.Depends...),
		Depends: append([]string{}, task.Depends...),
	}
	if _, exists := d.edges[task.Name]; !exists {
		d.edges[task.Name] = []string{}
	}
	for _, dep := range task.Depends {
		if _, exists := d.edges[dep]; !exists {
			d.edges[dep] = []string{}
		}
		d.edges[dep] = append(d.edges[dep], task.Name)
		if node, exists := d.nodes[dep]; exists {
			node.Depends = append(node.Depends, task.Name)
		}
	}
	return nil
}

// Validate checks for cycles and missing dependencies in the DAG.
func (d *DAG) Validate() error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	// Check for missing dependencies
	for _, node := range d.nodes {
		for _, dep := range node.Deps {
			if _, exists := d.nodes[dep]; !exists {
				return fmt.Errorf("dependency %s for task %s does not exist", dep, node.Name)
			}
		}
	}
	// Check for cycles
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	for node := range d.nodes {
		if !visited[node] {
			if err := d.detectCycle(node, visited, recStack); err != nil {
				return err
			}
		}
	}
	return nil
}

// detectCycle uses DFS to detect cycles in the DAG.
func (d *DAG) detectCycle(node string, visited, recStack map[string]bool) error {
	if _, exists := d.nodes[node]; !exists {
		return nil
	}
	if recStack[node] {
		return fmt.Errorf("cycle detected involving task: %s", node)
	}
	if visited[node] {
		return nil
	}
	visited[node] = true
	recStack[node] = true
	for _, neighbor := range d.nodes[node].Deps {
		if err := d.detectCycle(neighbor, visited, recStack); err != nil {
			return err
		}
	}
	recStack[node] = false
	return nil
}

// GetTopologicalOrder returns tasks in topological order using Kahn's algorithm.
func (d *DAG) GetTopologicalOrder() ([]*Task, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	// Check for missing dependencies
	for _, node := range d.nodes {
		for _, dep := range node.Deps {
			if _, exists := d.nodes[dep]; !exists {
				return nil, fmt.Errorf("dependency %s for task %s does not exist", dep, node.Name)
			}
		}
	}
	// Kahn's algorithm
	inDegree := make(map[string]int)
	for name := range d.nodes {
		inDegree[name] = 0
	}
	for _, node := range d.nodes {
		for range node.Deps {
			inDegree[node.Name]++
		}
	}
	queue := make([]string, 0)
	for name, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, name)
		}
	}
	order := make([]*Task, 0, len(d.nodes))
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		order = append(order, d.nodes[curr].Task)
		for _, dependent := range d.edges[curr] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}
	if len(order) != len(d.nodes) {
		return order, fmt.Errorf("cycle detected in DAG")
	}
	return order, nil
}

// GetDependencies returns all dependencies for a task.
func (d *DAG) GetDependencies(taskName string) ([]string, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	node, exists := d.nodes[taskName]
	if !exists {
		return nil, false
	}
	return append([]string{}, node.Deps...), true
}

// GetDependents returns all tasks that depend on a task.
func (d *DAG) GetDependents(taskName string) ([]string, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	node, exists := d.nodes[taskName]
	if !exists {
		return nil, false
	}
	return append([]string{}, node.Depends...), true
}
