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

func assertPodsLength(s *iapetus.Task) error {
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

func assertNsLength(s *iapetus.Task) error {
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
	teardownWorkflow := teardownWorkflow()
	if err := teardownWorkflow.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func setupWorkflow() iapetus.Workflow {
	return iapetus.Workflow{
		Name:     "Setup",
		LogLevel: 1,
		Steps: []iapetus.Task{
			{
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
			},
		},
	}
}

func teardownWorkflow() iapetus.Workflow {
	return iapetus.Workflow{
		Name:     "Cleanup",
		LogLevel: 1,
		Steps: []iapetus.Task{
			{
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
			},
		},
	}
}

func getWorkflowCasesForKubernetes() iapetus.Workflow {
	return iapetus.Workflow{
		Name: "Entire flow",
		PreRun: func(w *iapetus.Workflow) error {
			setupWorkflowSpec := setupWorkflow()
			if err := setupWorkflowSpec.Run(); err != nil {
				return err
			}
			return nil
		},
		LogLevel: 1,
		Steps: []iapetus.Task{
			{
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
			},
			{
				Name:    "kubectl-get-ns",
				Command: "kubectl",
				Args:    []string{"get", "ns", ns, "-o", "json"},
				Env:     []string{},
				Expected: iapetus.Output{
					ExitCode: 0,
				},
				Asserts: []func(*iapetus.Task) error{
					iapetus.AssertByExitCode,
					assertNsLength,
				},
			},
			{
				Name:    "kubectl-get-pods",
				Command: "kubectl",
				Args:    []string{"get", "pods", "-n", ns, "-o", "json"},
				Env:     []string{},
				Expected: iapetus.Output{
					ExitCode: 0,
				},
				Asserts: []func(*iapetus.Task) error{
					iapetus.AssertByExitCode,
					assertPodsLength,
				},
			},
			{
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
			},
			{
				Name:    "kubectl-get-pods-with-deployment",
				Command: "kubectl",
				Args:    []string{"get", "pods", "-n", ns, "-o", "json"},
				Env:     []string{},
				Retries: 1,
				Expected: iapetus.Output{
					ExitCode: 0,
				},
				Asserts: []func(*iapetus.Task) error{
					iapetus.AssertByExitCode,
					func(s *iapetus.Task) error {
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
					},
				},
			},
			{
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
			},
			{
				Name:    "kubectl-delete-ns",
				Command: "kubectl",
				Args:    []string{"delete", "ns", ns},
				Env:     []string{},
				Expected: iapetus.Output{
					ExitCode: 0,
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
				assertPodsLength,
			},
		},
	}
}
