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
