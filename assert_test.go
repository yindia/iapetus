package iapetus

import (
	"testing"
)

func TestAssertByExitCode(t *testing.T) {
	tests := []struct {
		name    string
		input   *Task
		wantErr bool
	}{
		{"Matching Exit Codes", &Task{Actual: Output{0, "", "", []string{}, []string{}}, Expected: Output{0, "", "", []string{}, []string{}}}, false},
		{"Mismatched Exit Codes", &Task{Actual: Output{1, "", "", []string{}, []string{}}, Expected: Output{0, "", "", []string{}, []string{}}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AssertByExitCode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("AssertByExitCode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAssertByOutputString(t *testing.T) {
	tests := []struct {
		name    string
		input   *Task
		wantErr bool
	}{
		{"Matching Output Strings", &Task{
			Actual:   Output{0, "output", "", []string{}, []string{}},
			Expected: Output{0, "output", "", []string{}, []string{}},
		}, false},
		{"Mismatched Output Strings", &Task{
			Actual:   Output{0, "output1", "", []string{}, []string{}},
			Expected: Output{0, "output2", "", []string{}, []string{}},
		}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AssertByOutputString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("AssertByOutputString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAssertByOutputJson(t *testing.T) {
	// Assuming parseJSONOutputs and compareJSON are implemented correctly
	tests := []struct {
		name    string
		input   *Task
		wantErr bool
	}{
		{"Matching JSON Outputs", &Task{
			Actual:   Output{0, `{"key":"value"}`, "", []string{}, []string{}},
			Expected: Output{0, `{"key":"value"}`, "", []string{}, []string{}},
		}, false},
		{"Mismatched JSON Outputs", &Task{
			Actual:   Output{0, `{"key":"value1"}`, "", []string{}, []string{}},
			Expected: Output{0, `{"key":"value2"}`, "", []string{}, []string{}},
		}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AssertByOutputJson(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("AssertByOutputJson() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAssertByError(t *testing.T) {
	tests := []struct {
		name    string
		input   *Task
		wantErr bool
	}{
		{"Matching Errors", &Task{
			Actual:   Output{0, "", "error", []string{}, []string{}},
			Expected: Output{0, "", "error", []string{}, []string{}},
		}, false},
		{"Mismatched Errors", &Task{
			Actual:   Output{0, "", "error1", []string{}, []string{}},
			Expected: Output{0, "", "error2", []string{}, []string{}},
		}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AssertByError(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("AssertByError() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAssertExitCode(t *testing.T) {
	// Implementation of TestAssertExitCode
}

func TestAssertOutputContains(t *testing.T) {
	// Implementation of TestAssertOutputContains
}

func TestAssertOutputEquals(t *testing.T) {
	// Implementation of TestAssertOutputEquals
}

func TestAssertOutputJsonEquals(t *testing.T) {
	// Implementation of TestAssertOutputJsonEquals
}

func TestAssertOutputMatchesRegexp(t *testing.T) {
	// Implementation of TestAssertOutputMatchesRegexp
}

func TestAssertExitCode_NewAPI(t *testing.T) {
	tests := []struct {
		name     string
		actual   int
		expected int
		wantErr  bool
	}{
		{"Match", 0, 0, false},
		{"Mismatch", 1, 0, true},
		{"Negative", -1, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Actual: Output{ExitCode: tt.actual}}
			err := AssertExitCode(tt.expected)(task)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAssertOutputContains_NewAPI(t *testing.T) {
	tests := []struct {
		name    string
		actual  string
		substr  string
		wantErr bool
	}{
		{"Contains", "hello foo bar", "foo", false},
		{"NotContains", "hello bar", "foo", true},
		{"Empty", "", "foo", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Actual: Output{Output: tt.actual}}
			err := AssertOutputContains(tt.substr)(task)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAssertOutputEquals_NewAPI(t *testing.T) {
	tests := []struct {
		name     string
		actual   string
		expected string
		wantErr  bool
	}{
		{"Equal", "foo", "foo", false},
		{"NotEqual", "foo", "bar", true},
		{"TrimSpace", " foo ", "foo", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Actual: Output{Output: tt.actual}}
			err := AssertOutputEquals(tt.expected)(task)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAssertOutputMatchesRegexp_NewAPI(t *testing.T) {
	tests := []struct {
		name    string
		actual  string
		pattern string
		wantErr bool
	}{
		{"Match", "foo123", `foo\d+`, false},
		{"NoMatch", "bar", `foo\d+`, true},
		{"InvalidPattern", "foo", `foo[`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Actual: Output{Output: tt.actual}}
			err := AssertOutputMatchesRegexp(tt.pattern)(task)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Note: For AssertOutputJsonEquals, a simple test for valid/invalid JSON and equality
func TestAssertOutputJsonEquals_NewAPI(t *testing.T) {
	tests := []struct {
		name     string
		actual   string
		expected string
		wantErr  bool
	}{
		{"EqualJSON", `{"foo":1}`, `{"foo":1}`, false},
		{"NotEqualJSON", `{"foo":1}`, `{"foo":2}`, true},
		{"InvalidActual", `notjson`, `{"foo":1}`, true},
		{"InvalidExpected", `{"foo":1}`, `notjson`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Actual: Output{Output: tt.actual}}
			err := AssertOutputJsonEquals(tt.expected)(task)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
