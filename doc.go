// Package iapetus provides a robust, extensible workflow engine for orchestrating and validating
// command-line tasks with parallel DAG execution, built-in assertions, and advanced observability.
//
// Features:
//   - Parallel, dependency-aware workflow execution (DAG-based)
//   - Built-in and custom assertions for validating outputs
//   - Retries, timeouts, and observability hooks for reliability and extensibility
//   - Fluent builder pattern for readable workflow and task construction
//   - Observability hooks and pluggable logging (zap)
//   - Container-ready: support for specifying images and environment variables
//   - Battle-tested: stress, property-based, and concurrency tests
//
// Backend Plugins:
//
// iapetus supports pluggable backends for task execution. A backend is any type that implements the Backend interface:
//
//	type Backend interface {
//	    RunTask(task *Task) error
//	    ValidateTask(task *Task) error
//	}
//
// To add a custom backend, implement this interface and register it with:
//
//	iapetus.RegisterBackend("my-backend", myBackendImpl)
//
// You can then set the backend for a workflow or individual task:
//
//	workflow.Backend = "my-backend" // sets default for all tasks in the workflow
//	task.SetBackend("my-backend")    // overrides for a specific task
//
// Built-in backends include "bash" (default) and "docker" (for containerized execution).
//
// Example custom backend:
//
//	type MyBackend struct{}
//	func (b *MyBackend) RunTask(task *iapetus.Task) error { /* ... */ return nil }
//	func (b *MyBackend) ValidateTask(task *iapetus.Task) error { return nil }
//
//	func init() {
//	    iapetus.RegisterBackend("my-backend", &MyBackend{})
//	}
//
// Example usage:
//
//	workflow := iapetus.NewWorkflow("example", zap.NewNop()).
//	    AddTask(*iapetus.NewTask("step1", 5*time.Second, nil).
//	        AddCommand("echo").
//	        AddArgs("hello").
//	        AssertExitCode(0).
//	        AssertOutputContains("hello"))
//
//	// Use AddOnTaskStartHook, AddOnTaskSuccessHook, etc. for observability and extensibility.
//
//	if err := workflow.Run(); err != nil {
//	    log.Fatalf("Workflow failed: %v", err)
//	}
//
// See the README for full documentation and examples.
package iapetus
