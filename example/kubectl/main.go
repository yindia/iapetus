package main

import (
	"fmt"
	"os"

	"github.com/yindia/iapetus"
)

func main() {
	tests := getTestCasesKubernetes("kubectl")
	for _, test := range tests {
		if err := test.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func getTestCasesKubernetes(binary string) []iapetus.Step {
	return []iapetus.Step{
		{
			Name:    "kubectl-get-pods",
			Command: binary,
			Args:    []string{"get", "pods", "-n", "default"},
			Env:     []string{},
			Expected: iapetus.Output{
				ExitCode: 0,
			},
			Asserts: []func(*iapetus.Step) error{
				iapetus.AssertByExitCode,
			},
		},
		{
			Name:    "kubectl-get-pods-json",
			Command: binary,
			Args:    []string{"get", "pods", "-n", "default", "-o", "json"},
			Env:     []string{},
			Expected: iapetus.Output{
				ExitCode: 0,
			},
			Asserts: []func(*iapetus.Step) error{
				iapetus.AssertByExitCode,
				func(s *iapetus.Step) error {
					// convert string to pods specs 
					if s.Actual.Output == "" {
						return fmt.Errorf("output is empty")
					}
					return nil
				},
			},
		},
	}
}
