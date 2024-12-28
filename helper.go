package iapetus

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

func ReadFile(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return string(content)
}

func getExitCode(err error) int {
	if err == nil {
		return 0
	}
	if exitError, ok := err.(*exec.ExitError); ok {
		return exitError.ExitCode()
	}
	return -1
}

func parseJSONOutputs(actual, expected string) (interface{}, interface{}, error) {
	var actualData, expectedData interface{}

	if err := json.Unmarshal([]byte(actual), &actualData); err != nil {
		return nil, nil, fmt.Errorf("failed to parse actual output as JSON: %w", err)
	}

	if err := json.Unmarshal([]byte(expected), &expectedData); err != nil {
		return nil, nil, fmt.Errorf("failed to parse expected output as JSON: %w", err)
	}

	return actualData, expectedData, nil
}