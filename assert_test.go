package iapetus

import (
	"testing"
)

func TestAssertExitCode(t *testing.T) {
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

func TestAssertOutputContains(t *testing.T) {
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

func TestAssertOutputEquals(t *testing.T) {
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

func TestAssertOutputJsonEquals(t *testing.T) {
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

func TestAssertOutputMatchesRegexp(t *testing.T) {
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
