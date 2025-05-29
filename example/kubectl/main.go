package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/yindia/iapetus"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

const nsA = "test-a"
const nsB = "test-b"

var TASK_CREATE_KIND_CLUSTER = iapetus.Task{
	Name:    "Create Kind Cluster",
	Command: "kind",
	Args:    []string{"create", "cluster"},
	Asserts: []func(*iapetus.Task) error{
		iapetus.AssertExitCode(0),
	},
}

var TASK_CREATE_NS_A = iapetus.Task{
	Name:    "Create Namespace A",
	Command: "kubectl",
	Args:    []string{"create", "ns", nsA},
	Asserts: []func(*iapetus.Task) error{
		iapetus.AssertExitCode(0),
	},
	Depends: []string{"Create Kind Cluster"},
}

var TASK_CREATE_NS_B = iapetus.Task{
	Name:    "Create Namespace B",
	Command: "kubectl",
	Args:    []string{"create", "ns", nsB},
	Asserts: []func(*iapetus.Task) error{
		iapetus.AssertExitCode(0),
	},
	Depends: []string{"Create Kind Cluster"},
}

var TASK_DEPLOY_NGINX_A = iapetus.Task{
	Name:    "Deploy Nginx in A",
	Command: "kubectl",
	Args:    []string{"create", "deployment", "nginx-a", "--image", "nginx", "-n", nsA},
	Asserts: []func(*iapetus.Task) error{
		iapetus.AssertExitCode(0),
	},
	Depends: []string{"Create Namespace A"},
}

var TASK_DEPLOY_NGINX_B = iapetus.Task{
	Name:    "Deploy Nginx in B",
	Command: "kubectl",
	Args:    []string{"create", "deployment", "nginx-b", "--image", "nginx", "-n", nsB},
	Asserts: []func(*iapetus.Task) error{
		iapetus.AssertExitCode(0),
	},
	Depends: []string{"Create Namespace B"},
}

var TASK_GET_PODS_A = iapetus.Task{
	Name:    "Get Pods in A",
	Command: "kubectl",
	Args:    []string{"get", "pods", "-n", nsA, "-o", "json"},
	Asserts: []func(*iapetus.Task) error{
		iapetus.AssertExitCode(0),
		AssertPodsExist,
	},
	Depends: []string{"Deploy Nginx in A"},
}

var TASK_GET_PODS_B = iapetus.Task{
	Name:    "Get Pods in B",
	Command: "kubectl",
	Args:    []string{"get", "pods", "-n", nsB, "-o", "json"},
	Asserts: []func(*iapetus.Task) error{
		iapetus.AssertExitCode(0),
		AssertPodsExist,
	},
	Depends: []string{"Deploy Nginx in B"},
}

var TASK_DELETE_DEPLOYMENT_A = iapetus.Task{
	Name:    "Delete Deployment A",
	Command: "kubectl",
	Args:    []string{"delete", "deployment", "nginx-a", "-n", nsA},
	Asserts: []func(*iapetus.Task) error{
		iapetus.AssertExitCode(0),
	},
	Depends: []string{"Get Pods in A"},
}

var TASK_DELETE_DEPLOYMENT_B = iapetus.Task{
	Name:    "Delete Deployment B",
	Command: "kubectl",
	Args:    []string{"delete", "deployment", "nginx-b", "-n", nsB},
	Asserts: []func(*iapetus.Task) error{
		iapetus.AssertExitCode(0),
	},
	Depends: []string{"Get Pods in B"},
}

var TASK_DELETE_NS_A = iapetus.Task{
	Name:    "Delete Namespace A",
	Command: "kubectl",
	Args:    []string{"delete", "ns", nsA},
	Asserts: []func(*iapetus.Task) error{
		iapetus.AssertExitCode(0),
	},
	Depends: []string{"Delete Deployment A"},
}

var TASK_DELETE_NS_B = iapetus.Task{
	Name:    "Delete Namespace B",
	Command: "kubectl",
	Args:    []string{"delete", "ns", nsB},
	Asserts: []func(*iapetus.Task) error{
		iapetus.AssertExitCode(0),
	},
	Depends: []string{"Delete Deployment B"},
}

var TASK_DELETE_KIND_CLUSTER = iapetus.Task{
	Name:    "Delete Kind Cluster",
	Command: "kind",
	Args:    []string{"delete", "cluster"},
	Asserts: []func(*iapetus.Task) error{
		iapetus.AssertExitCode(0),
	},
	Depends: []string{"Delete Namespace A", "Delete Namespace B"},
}

// Custom assertion: check namespace exists in output
func AssertNamespaceExists(ns string) func(*iapetus.Task) error {
	return func(t *iapetus.Task) error {
		if !json.Valid([]byte(t.Actual.Output)) {
			return fmt.Errorf("output is not valid JSON")
		}
		var nsObj v1.Namespace
		if err := json.Unmarshal([]byte(t.Actual.Output), &nsObj); err != nil {
			return fmt.Errorf("failed to unmarshal namespace: %w", err)
		}
		if nsObj.Name != ns {
			return fmt.Errorf("namespace name mismatch: expected %s, got %s", ns, nsObj.Name)
		}
		return nil
	}
}

// Custom assertion: check deployment exists in output
func AssertDeploymentExists(depName string) func(*iapetus.Task) error {
	return func(t *iapetus.Task) error {
		if !json.Valid([]byte(t.Actual.Output)) {
			return fmt.Errorf("output is not valid JSON")
		}
		var depList appsv1.DeploymentList
		if err := json.Unmarshal([]byte(t.Actual.Output), &depList); err != nil {
			return fmt.Errorf("failed to unmarshal deployment list: %w", err)
		}
		for _, dep := range depList.Items {
			if dep.Name == depName {
				return nil
			}
		}
		return fmt.Errorf("deployment %s not found", depName)
	}
}

// Custom assertion: check deployment is deleted (no deployments)
func AssertNoDeployments() func(*iapetus.Task) error {
	return func(t *iapetus.Task) error {
		if !json.Valid([]byte(t.Actual.Output)) {
			return fmt.Errorf("output is not valid JSON")
		}
		var depList appsv1.DeploymentList
		if err := json.Unmarshal([]byte(t.Actual.Output), &depList); err != nil {
			return fmt.Errorf("failed to unmarshal deployment list: %w", err)
		}
		if len(depList.Items) != 0 {
			return fmt.Errorf("expected no deployments, found %d", len(depList.Items))
		}
		return nil
	}
}

// Additional tasks for advanced assertions and dependencies
var TASK_GET_NS_A = iapetus.Task{
	Name:    "Get Namespace A",
	Command: "kubectl",
	Args:    []string{"get", "ns", nsA, "-o", "json"},
	Asserts: []func(*iapetus.Task) error{
		iapetus.AssertExitCode(0),
		AssertNamespaceExists(nsA),
	},
	Depends: []string{"Create Namespace A"},
}

var TASK_GET_DEPLOYMENT_A = iapetus.Task{
	Name:    "Get Deployment A",
	Command: "kubectl",
	Args:    []string{"get", "deployment", "-n", nsA, "-o", "json"},
	Asserts: []func(*iapetus.Task) error{
		iapetus.AssertExitCode(0),
		AssertDeploymentExists("nginx-a"),
	},
	Depends: []string{"Deploy Nginx in A"},
}

var TASK_GET_DEPLOYMENT_B = iapetus.Task{
	Name:    "Get Deployment B",
	Command: "kubectl",
	Args:    []string{"get", "deployment", "-n", nsB, "-o", "json"},
	Asserts: []func(*iapetus.Task) error{
		iapetus.AssertExitCode(0),
		AssertDeploymentExists("nginx-b"),
	},
	Depends: []string{"Deploy Nginx in B"},
}

var TASK_CHECK_NO_DEPLOYMENT_A = iapetus.Task{
	Name:    "Check No Deployment A",
	Command: "kubectl",
	Args:    []string{"get", "deployment", "-n", nsA, "-o", "json"},
	Asserts: []func(*iapetus.Task) error{
		iapetus.AssertExitCode(0),
		AssertNoDeployments(),
	},
	Depends: []string{"Delete Deployment A"},
}

var TASK_CHECK_NO_DEPLOYMENT_B = iapetus.Task{
	Name:    "Check No Deployment B",
	Command: "kubectl",
	Args:    []string{"get", "deployment", "-n", nsB, "-o", "json"},
	Asserts: []func(*iapetus.Task) error{
		iapetus.AssertExitCode(0),
		AssertNoDeployments(),
	},
	Depends: []string{"Delete Deployment B"},
}

// Final summary task that depends on both pod checks
var TASK_SUMMARY = iapetus.Task{
	Name:    "Summary",
	Command: "echo",
	Args:    []string{"All checks passed!"},
	Asserts: []func(*iapetus.Task) error{
		iapetus.AssertExitCode(0),
		// Could add more summary logic here
	},
	Depends: []string{"Get Pods in A", "Get Pods in B"},
}

func AssertPodsExist(t *iapetus.Task) error {
	pods := &v1.PodList{}
	err := json.Unmarshal([]byte(t.Actual.Output), &pods)
	if err != nil {
		return fmt.Errorf("failed to unmarshal pods: %w\nRaw output: %s", err, t.Actual.Output)
	}
	if len(pods.Items) == 0 {
		return fmt.Errorf("no pods found\nRaw output: %s", t.Actual.Output)
	}
	return nil
}

func main() {
	// Build the DAG workflow with advanced assertions and dependencies
	workflow := iapetus.NewWorkflow("K8s DAG Example", zap.NewNop())

	// Register extensibility hooks for observability
	workflow.AddOnTaskStartHook(func(task *iapetus.Task) {
		fmt.Printf("[HOOK] Task started: %s\n", task.Name)
	})
	workflow.AddOnTaskSuccessHook(func(task *iapetus.Task) {
		fmt.Printf("[HOOK] Task succeeded: %s\n", task.Name)
	})
	workflow.AddOnTaskFailureHook(func(task *iapetus.Task, err error) {
		fmt.Printf("[HOOK] Task failed: %s, error: %v\n", task.Name, err)
	})
	workflow.AddOnTaskCompleteHook(func(task *iapetus.Task) {
		fmt.Printf("[HOOK] Task completed: %s\n", task.Name)
	})

	workflow.Steps = []iapetus.Task{
		TASK_CREATE_KIND_CLUSTER,
		TASK_CREATE_NS_A,
		TASK_CREATE_NS_B,
		TASK_GET_NS_A,
		TASK_DEPLOY_NGINX_A,
		TASK_DEPLOY_NGINX_B,
		TASK_GET_DEPLOYMENT_A,
		TASK_GET_DEPLOYMENT_B,
		TASK_GET_PODS_A,
		TASK_GET_PODS_B,
		TASK_DELETE_DEPLOYMENT_A,
		TASK_DELETE_DEPLOYMENT_B,
		TASK_CHECK_NO_DEPLOYMENT_A,
		TASK_CHECK_NO_DEPLOYMENT_B,
		TASK_DELETE_NS_A,
		TASK_DELETE_NS_B,
		TASK_DELETE_KIND_CLUSTER,
		TASK_SUMMARY,
	}

	fmt.Println("--- Running Kubernetes DAG Workflow with Advanced Assertions ---")
	err := workflow.Run()
	if err != nil {
		fmt.Println("Workflow failed:", err)
		os.Exit(1)
	}
	fmt.Println("Workflow completed successfully!")
}
