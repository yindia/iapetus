package iapetus

import (
	"errors"
	"os/exec"
	"testing"
)

func TestGetExitCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{"No error", nil, 0},
		{"Exit error", &exec.ExitError{}, -1},
		{"Non-exit error", errors.New("some error"), -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getExitCode(tt.err); got != tt.expected {
				t.Errorf("getExitCode() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseJSONOutputs(t *testing.T) {
	tests := []struct {
		name             string
		actual, expected string
		wantErr          bool
	}{
		{"Valid JSON", `{"key": "value"}`, `{"key": "value"}`, false},
		{"Invalid actual JSON", `{"key": "value"`, `{"key": "value"}`, true},
		{"Invalid expected JSON", `{"key": "value"}`, `{"key": "value"`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := parseJSONOutputs(tt.actual, tt.expected)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseJSONOutputs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}