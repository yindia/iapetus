package iapetus

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"testing/quick"

	"github.com/stretchr/testify/assert"
)

func TestDAG_AddTask(t *testing.T) {
	t.Run("add single task", func(t *testing.T) {
		dag := NewDag()
		task := &Task{Name: "task1"}
		err := dag.AddTask(task)
		assert.NoError(t, err)
	})

	t.Run("add duplicate task", func(t *testing.T) {
		dag := NewDag()
		task := &Task{Name: "task1"}
		_ = dag.AddTask(task)
		err := dag.AddTask(task)
		assert.Error(t, err)
	})

	t.Run("add task with missing dependency", func(t *testing.T) {
		dag := NewDag()
		task := &Task{Name: "task2", Depends: []string{"notfound"}}
		err := dag.AddTask(task)
		assert.NoError(t, err)
		err = dag.Validate()
		assert.Error(t, err)
	})

	t.Run("add task with valid dependency", func(t *testing.T) {
		dag := NewDag()
		task1 := &Task{Name: "task1"}
		task2 := &Task{Name: "task2", Depends: []string{"task1"}}
		assert.NoError(t, dag.AddTask(task1))
		assert.NoError(t, dag.AddTask(task2))
		assert.NoError(t, dag.Validate())
	})
}

func TestDAG_Validate(t *testing.T) {
	dag := NewDag()
	task1 := &Task{Name: "task1"}
	task2 := &Task{Name: "task2", Depends: []string{"task1"}}
	task3 := &Task{Name: "task3", Depends: []string{"task2"}}
	assert.NoError(t, dag.AddTask(task1))
	assert.NoError(t, dag.AddTask(task2))
	assert.NoError(t, dag.AddTask(task3))
	assert.NoError(t, dag.Validate())

	t.Run("cycle detection", func(t *testing.T) {
		dagWithCycle := NewDag()
		t1 := &Task{Name: "a", Depends: []string{"c"}}
		t2 := &Task{Name: "b", Depends: []string{"a"}}
		t3 := &Task{Name: "c", Depends: []string{"b"}}
		assert.NoError(t, dagWithCycle.AddTask(&Task{Name: "root"}))
		assert.NoError(t, dagWithCycle.AddTask(t1))
		assert.NoError(t, dagWithCycle.AddTask(t2))
		assert.NoError(t, dagWithCycle.AddTask(t3))
		err := dagWithCycle.Validate()
		assert.Error(t, err)
	})
}

func TestDAG_GetTopologicalOrder(t *testing.T) {
	dag := NewDag()
	tasks := []*Task{
		{Name: "task1"},
		{Name: "task2", Depends: []string{"task1"}},
		{Name: "task3", Depends: []string{"task2"}},
	}
	for _, task := range tasks {
		assert.NoError(t, dag.AddTask(task))
	}
	order, err := dag.GetTopologicalOrder()
	assert.NoError(t, err)
	assert.Equal(t, 3, len(order))
	// order[0] should be task1, order[1] task2, order[2] task3
	assert.Equal(t, "task1", order[0].Name)
	assert.Equal(t, "task2", order[1].Name)
	assert.Equal(t, "task3", order[2].Name)
}

func TestDAG_GetDependencies(t *testing.T) {
	dag := NewDag()
	task1 := &Task{Name: "task1"}
	task2 := &Task{Name: "task2", Depends: []string{"task1"}}
	assert.NoError(t, dag.AddTask(task1))
	assert.NoError(t, dag.AddTask(task2))
	deps, ok := dag.GetDependencies("task2")
	assert.True(t, ok)
	assert.Equal(t, []string{"task1"}, deps)
	_, ok = dag.GetDependencies("notfound")
	assert.False(t, ok)
}

func TestDAG_GetDependents(t *testing.T) {
	dag := NewDag()
	task1 := &Task{Name: "task1"}
	task2 := &Task{Name: "task2", Depends: []string{"task1"}}
	assert.NoError(t, dag.AddTask(task1))
	assert.NoError(t, dag.AddTask(task2))
	deps, ok := dag.GetDependents("task1")
	assert.True(t, ok)
	assert.Contains(t, deps, "task2")
	_, ok = dag.GetDependents("notfound")
	assert.False(t, ok)
}

func TestDAG_ConcurrentOperations(t *testing.T) {
	dag := NewDag()
	n := 100
	tasks := make([]*Task, n)
	for i := 0; i < n; i++ {
		tasks[i] = &Task{Name: fmt.Sprintf("task%d", i)}
		if i > 0 {
			tasks[i].Depends = []string{fmt.Sprintf("task%d", i-1)}
		}
	}
	var wg sync.WaitGroup
	for _, task := range tasks {
		wg.Add(1)
		go func(task *Task) {
			defer wg.Done()
			assert.NoError(t, dag.AddTask(task))
		}(task)
	}
	wg.Wait()
	// Validate DAG after concurrent adds
	assert.NoError(t, dag.Validate())

	// Concurrent dependency queries
	wg = sync.WaitGroup{}
	for _, task := range tasks {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			_, _ = dag.GetDependencies(name)
			_, _ = dag.GetDependents(name)
		}(task.Name)
	}
	wg.Wait()
}

func TestDAG_DiamondDependency(t *testing.T) {
	dag := NewDag()
	tasks := []*Task{
		{Name: "A"},
		{Name: "B", Depends: []string{"A"}},
		{Name: "C", Depends: []string{"A"}},
		{Name: "D", Depends: []string{"B", "C"}},
	}
	for _, task := range tasks {
		assert.NoError(t, dag.AddTask(task))
	}
	order, err := dag.GetTopologicalOrder()
	assert.NoError(t, err)
	// D must come after B and C, which both come after A
	pos := map[string]int{}
	for i, task := range order {
		pos[task.Name] = i
	}
	assert.True(t, pos["A"] < pos["B"] && pos["A"] < pos["C"])
	assert.True(t, pos["B"] < pos["D"] && pos["C"] < pos["D"])
}

func TestDAG_MultipleRootsAndLeaves(t *testing.T) {
	dag := NewDag()
	tasks := []*Task{
		{Name: "root1"},
		{Name: "root2"},
		{Name: "mid1", Depends: []string{"root1"}},
		{Name: "mid2", Depends: []string{"root2"}},
		{Name: "leaf1", Depends: []string{"mid1"}},
		{Name: "leaf2", Depends: []string{"mid2"}},
	}
	for _, task := range tasks {
		assert.NoError(t, dag.AddTask(task))
	}
	order, err := dag.GetTopologicalOrder()
	assert.NoError(t, err)
	// All tasks should be present
	assert.Equal(t, 6, len(order))
}

func TestDAG_SelfDependency(t *testing.T) {
	dag := NewDag()
	task := &Task{Name: "self", Depends: []string{"self"}}
	assert.NoError(t, dag.AddTask(task))
	assert.Error(t, dag.Validate())
}

func TestDAG_LongChain(t *testing.T) {
	dag := NewDag()
	const n = 1000
	for i := 0; i < n; i++ {
		task := &Task{Name: fmt.Sprintf("t%d", i)}
		if i > 0 {
			task.Depends = []string{fmt.Sprintf("t%d", i-1)}
		}
		assert.NoError(t, dag.AddTask(task))
	}
	assert.NoError(t, dag.Validate())
	order, err := dag.GetTopologicalOrder()
	assert.NoError(t, err)
	assert.Equal(t, n, len(order))
	// Ensure order is correct
	for i := 1; i < n; i++ {
		prev := fmt.Sprintf("t%d", i-1)
		curr := fmt.Sprintf("t%d", i)
		pos := map[string]int{}
		for j, task := range order {
			pos[task.Name] = j
		}
		assert.True(t, pos[prev] < pos[curr])
	}
}

func TestDAG_LargeGraphStress(t *testing.T) {
	dag := NewDag()
	const n = 5000
	for i := 0; i < n; i++ {
		task := &Task{Name: fmt.Sprintf("t%d", i)}
		if i > 0 {
			task.Depends = []string{fmt.Sprintf("t%d", i-1)}
		}
		assert.NoError(t, dag.AddTask(task))
	}
	assert.NoError(t, dag.Validate())
}

func TestDAG_AddTaskEdgeCases(t *testing.T) {
	dag := NewDag()
	t.Run("nil task", func(t *testing.T) {
		var nilTask *Task
		err := dag.AddTask(nilTask)
		assert.Error(t, err)
	})
	t.Run("empty name", func(t *testing.T) {
		err := dag.AddTask(&Task{Name: ""})
		assert.Error(t, err)
	})
	t.Run("duplicate dependencies", func(t *testing.T) {
		task1 := &Task{Name: "a"}
		task2 := &Task{Name: "b", Depends: []string{"a", "a"}}
		assert.NoError(t, dag.AddTask(task1))
		assert.NoError(t, dag.AddTask(task2))
		// Should not error, but dependencies should be unique in logic
		deps, ok := dag.GetDependencies("b")
		assert.True(t, ok)
		assert.Equal(t, []string{"a", "a"}, deps)
	})
}

func TestDAG_ValidateEmptyDAG(t *testing.T) {
	dag := NewDag()
	assert.NoError(t, dag.Validate())
}

func TestDAG_TopologicalOrderOnCycle(t *testing.T) {
	dag := NewDag()
	t1 := &Task{Name: "a", Depends: []string{"b"}}
	t2 := &Task{Name: "b", Depends: []string{"a"}}
	assert.NoError(t, dag.AddTask(t1))
	assert.NoError(t, dag.AddTask(t2))
	assert.Error(t, dag.Validate())
	order, err := dag.GetTopologicalOrder()
	assert.Error(t, err) // Now expect error due to cycle
	// Optionally, check that order is not valid (both present)
	if len(order) == 2 {
		pos := map[string]int{}
		for i, task := range order {
			pos[task.Name] = i
		}
		assert.NotEqual(t, pos["a"], pos["b"]) // Both present, but order is ambiguous
	}
}

func TestDAG_TopologicalOrderCorrectness(t *testing.T) {
	dag := NewDag()
	tasks := []*Task{
		{Name: "A"},
		{Name: "B", Depends: []string{"A"}},
		{Name: "C", Depends: []string{"A"}},
		{Name: "D", Depends: []string{"B", "C"}},
	}
	for _, task := range tasks {
		assert.NoError(t, dag.AddTask(task))
	}
	order, err := dag.GetTopologicalOrder()
	assert.NoError(t, err)
	pos := map[string]int{}
	for i, task := range order {
		pos[task.Name] = i
	}
	for _, task := range tasks {
		for _, dep := range task.Depends {
			assert.True(t, pos[dep] < pos[task.Name], "dependency %s should come before %s", dep, task.Name)
		}
	}
}

func TestDAG_ConcurrentAddAndValidate(t *testing.T) {
	dag := NewDag()
	const n = 100
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			task := &Task{Name: fmt.Sprintf("t%d", i)}
			if i > 0 {
				task.Depends = []string{fmt.Sprintf("t%d", i-1)}
			}
			assert.NoError(t, dag.AddTask(task))
		}(i)
	}
	wg.Wait()
	assert.NoError(t, dag.Validate())
}

func TestDAG_ConcurrentTopologicalSort(t *testing.T) {
	dag := NewDag()
	const n = 100
	for i := 0; i < n; i++ {
		task := &Task{Name: fmt.Sprintf("t%d", i)}
		if i > 0 {
			task.Depends = []string{fmt.Sprintf("t%d", i-1)}
		}
		assert.NoError(t, dag.AddTask(task))
	}
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			order, err := dag.GetTopologicalOrder()
			assert.NoError(t, err)
			assert.Equal(t, n, len(order))
		}()
	}
	wg.Wait()
}

// Property-based test: For any valid DAG, topological order is valid and complete
func TestDAG_TopologicalOrder_Property(t *testing.T) {
	f := func(numTasks uint8) bool {
		n := int(numTasks%20) + 2 // 2-21 tasks

		dag := NewDag()
		names := make([]string, n)
		for i := 0; i < n; i++ {
			names[i] = fmt.Sprintf("t%d", i)
		}
		// Randomly assign dependencies (no cycles)
		for i := 0; i < n; i++ {
			deps := []string{}
			for j := 0; j < i; j++ {
				if rand.Float64() < 0.3 {
					deps = append(deps, names[j])
				}
			}
			_ = dag.AddTask(&Task{Name: names[i], Depends: deps})
		}
		if err := dag.Validate(); err != nil {
			return true // skip invalid DAGs
		}
		order, err := dag.GetTopologicalOrder()
		if err != nil {
			t.Logf("unexpected error: %v", err)
			return false
		}
		// Check all tasks present
		if len(order) != n {
			t.Logf("order len %d != n %d", len(order), n)
			return false
		}
		// Check dependencies respected
		pos := map[string]int{}
		for i, task := range order {
			pos[task.Name] = i
		}
		for i := 0; i < n; i++ {
			for _, dep := range dag.nodes[names[i]].Deps {
				if pos[dep] > pos[names[i]] {
					t.Logf("dependency %s after %s", dep, names[i])
					return false
				}
			}
		}
		return true
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}
