package main

import (
	"encoding/json"
	"fmt"
	"os"

	appsv1 "k8s.io/api/apps/v1"

	"github.com/yindia/iapetus"
	v1 "k8s.io/api/core/v1"
)

const ns = "test"

var TASK_CREATE_KIND_CLUSTER = iapetus.Task{
	Name:    "Create Kind Cluster",
	Command: "kind",
	Args:    []string{"create", "cluster"},
	Env:     []string{},
	Expected: iapetus.Output{
		ExitCode: 0,
	},
	Asserts: []func(*iapetus.Task) error{
		iapetus.AssertByExitCode,
	},
}

var TASK_DELETE_KIND_CLUSTER = iapetus.Task{
	Name:    "Delete Kind Cluster",
	Command: "kind",
	Args:    []string{"delete", "cluster"},
	Env:     []string{},
	Expected: iapetus.Output{
		ExitCode: 0,
	},
	Asserts: []func(*iapetus.Task) error{
		iapetus.AssertByExitCode,
	},
}

var TASK_CREATE_NGINX_DEPLOYMENT = iapetus.Task{
	Name:    "kubectl-create-deployment",
	Command: "kubectl",
	Args:    []string{"create", "deployment", "test", "--image", "nginx", "--replicas", "30", "-n", ns},
	Env:     []string{},
	Expected: iapetus.Output{
		ExitCode: 0,
	},
	Asserts: []func(*iapetus.Task) error{
		iapetus.AssertByExitCode,
	},
}

var TASK_DELETE_NGINX_DEPLOYMENT = iapetus.Task{
	Name:    "kubectl-delete-deployment",
	Command: "kubectl",
	Args:    []string{"delete", "deployment", "test", "-n", ns},
	Env:     []string{},
	Expected: iapetus.Output{
		ExitCode: 0,
	},
	Asserts: []func(*iapetus.Task) error{
		iapetus.AssertByExitCode,
	},
}

var TASK_DELETE_NS = iapetus.Task{
	Name:    "kubectl-delete-ns",
	Command: "kubectl",
	Args:    []string{"delete", "ns", ns},
	Env:     []string{},
	Expected: iapetus.Output{
		ExitCode: 0,
	},
}

var TASK_CREATE_NS = iapetus.Task{
	Name:    "kubectl-create-ns",
	Command: "kubectl",
	Args:    []string{"create", "ns", ns},
	Env:     []string{},
	Expected: iapetus.Output{
		ExitCode: 0,
	},
	Asserts: []func(*iapetus.Task) error{
		iapetus.AssertByExitCode,
	},
}

func AssertByCustomDeploymentCheck(s *iapetus.Task) error {
	deployment := &appsv1.DeploymentList{}
	err := json.Unmarshal([]byte(s.Actual.Output), &deployment)
	if err != nil {
		return fmt.Errorf("failed to unmarshal deployment specs: %w", err)
	}
	if len(deployment.Items) == 1 {
		return fmt.Errorf("deployment length should be 1")
	}
	for _, item := range deployment.Items {
		if item.Name == "test" {
			for _, container := range item.Spec.Template.Spec.Containers {
				if container.Image != "nginx" {
					return fmt.Errorf("container image should be nginx")
				}
			}
			if item.Status.Replicas != *item.Spec.Replicas {
				return fmt.Errorf("deployment replicas do not match desired state")
			}
		}
	}
	return nil
}

func AssertByPodsCount(s *iapetus.Task) error {
	pods := &v1.PodList{}
	err := json.Unmarshal([]byte(s.Actual.Output), &pods)
	if err != nil {
		return fmt.Errorf("failed to unmarshal pods specs: %w", err)
	}
	if len(pods.Items) > 0 {
		return fmt.Errorf("pods length should be 0")
	}
	return nil
}

func AssertByNsCount(s *iapetus.Task) error {
	ns := &v1.Namespace{}
	err := json.Unmarshal([]byte(s.Actual.Output), &ns)
	if err != nil {
		return fmt.Errorf("failed to unmarshal ns specs: %w", err)
	}
	if ns.Name != "test" {
		return fmt.Errorf("ns name should be test")
	}
	return nil
}

func main() {
	workflow := getWorkflowCasesForKubernetes()
	if err := workflow.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	tests := getTestCasesForKubernetes()
	for _, test := range tests {
		if err := test.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if err := TASK_DELETE_KIND_CLUSTER.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getWorkflowCasesForKubernetes() iapetus.Workflow {
	return iapetus.Workflow{
		Name: "Entire flow",
		PreRun: func(w *iapetus.Workflow) error {
			return TASK_CREATE_KIND_CLUSTER.Run()
		},
		LogLevel: 1,
		PostRun: func(w *iapetus.Workflow) error {
			return TASK_DELETE_NS.Run()
		},
		Steps: []iapetus.Task{
			{
				Name:    "kubectl-get-ns",
				Command: "kubectl",
				Args:    []string{"get", "ns", ns, "-o", "json"},
				Env:     []string{},
				PreRun: func(w *iapetus.Task) error {
					return TASK_CREATE_NS.Run()
				},
				Expected: iapetus.Output{
					ExitCode: 0,
				},
				Asserts: []func(*iapetus.Task) error{
					iapetus.AssertByExitCode,
					AssertByNsCount,
				},
			},
			{
				Name:    "kubectl-get-pods-with-deployment",
				Command: "kubectl",
				Args:    []string{"get", "pods", "-n", ns, "-o", "json"},
				Env:     []string{},
				Retries: 1,
				PreRun: func(w *iapetus.Task) error {
					return TASK_CREATE_NGINX_DEPLOYMENT.Run()
				},
				PostRun: func(w *iapetus.Task) error {
					return TASK_DELETE_NGINX_DEPLOYMENT.Run()
				},
				Expected: iapetus.Output{
					ExitCode: 0,
				},
				Asserts: []func(*iapetus.Task) error{
					iapetus.AssertByExitCode,
					AssertByCustomDeploymentCheck,
				},
			},
		},
	}
}

func getTestCasesForKubernetes() []iapetus.Task {
	return []iapetus.Task{
		{
			Name:    "kubectl-get-pods",
			Command: "kubectl",
			Args:    []string{"get", "pods", "-n", "default"},
			Env:     []string{},
			PreRun: func(t *iapetus.Task) error {
				fmt.Println("Task Started, I can create a temp file before running a command")
				return nil
			},
			Expected: iapetus.Output{
				ExitCode: 0,
			},
			Asserts: []func(*iapetus.Task) error{
				iapetus.AssertByExitCode,
			},
		},
		{
			Name:    "kubectl-get-pods-json",
			Command: "kubectl",
			Args:    []string{"get", "pods", "-n", "default", "-o", "json"},
			Env:     []string{},
			Expected: iapetus.Output{
				ExitCode: 0,
			},
			Asserts: []func(*iapetus.Task) error{
				iapetus.AssertByExitCode,
				AssertByPodsCount,
			},
		},
	}
}
