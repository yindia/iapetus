package iapetus

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

// ReadFile reads the entire contents of a file at the given path and returns it as a string.
// It panics if the file cannot be read.
func ReadFile(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return string(content)
}

// getExitCode extracts the exit code from an error that may be an *exec.ExitError.
// Returns 0 if err is nil, the actual exit code if err is an *exec.ExitError,
// or -1 for other error types.
func getExitCode(err error) int {
	if err == nil {
		return 0
	}
	if exitError, ok := err.(*exec.ExitError); ok {
		return exitError.ExitCode()
	}
	return -1
}

// parseJSONOutputs unmarshals two JSON strings into interface{} values for comparison.
// Returns the parsed actual and expected values, or an error if either string
// cannot be parsed as valid JSON.
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
