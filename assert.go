package iapetus

import (
	"fmt"
	"regexp"
	"strings"

	jd "github.com/josephburnett/jd/lib"
)

// Aggregate assertion errors
// AssertionErrors collects multiple assertion failures
// Implements error interface
type AssertionErrors []error

func (ae AssertionErrors) Error() string {
	var msgs []string
	for _, err := range ae {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// RunAssertions runs all assertions and aggregates errors
func RunAssertions(task *Task) error {
	var errs AssertionErrors
	for _, assert := range task.Asserts {
		if err := assert(task); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

// Output normalization helper
func normalizeOutput(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(s, "\r\n", "\n"))
}

// New assertion functions that accept expected value and return a closure
// These are preferred for new code

// AssertExitCode returns an assertion that checks the exit code
func AssertExitCode(expected int) func(*Task) error {
	return func(i *Task) error {
		if i.Actual.ExitCode != expected {
			return fmt.Errorf("exit code mismatch: expected %d, got %d", expected, i.Actual.ExitCode)
		}
		return nil
	}
}

// AssertOutputContains returns an assertion that checks if output contains a substring
func AssertOutputContains(substr string) func(*Task) error {
	return func(i *Task) error {
		if !strings.Contains(i.Actual.Output, substr) {
			return fmt.Errorf("output does not contain expected substring: %q", substr)
		}
		return nil
	}
}

// AssertOutputEquals returns an assertion that checks if output matches exactly
func AssertOutputEquals(expected string) func(*Task) error {
	return func(i *Task) error {
		actual := normalizeOutput(i.Actual.Output)
		exp := normalizeOutput(expected)
		if actual != exp {
			return fmt.Errorf("output mismatch: expected %q, got %q", exp, actual)
		}
		return nil
	}
}

// AssertOutputJsonEquals returns an assertion that checks if output JSON matches expected JSON
func AssertOutputJsonEquals(expected string, skipJsonNodes ...string) func(*Task) error {
	return func(i *Task) error {
		expectation, err := jd.ReadJsonString(expected)
		if err != nil {
			return fmt.Errorf("failed to read expectation: %w", err)
		}
		parsedOutput, err := jd.ReadJsonString(normalizeOutput(i.Actual.Output))
		if err != nil {
			return fmt.Errorf("failed to parse output: %w", err)
		}
		diff := expectation.Diff(parsedOutput)
		var errors []string
		for _, d := range diff {
			if shouldSkipPath(d.Path, skipJsonNodes) {
				continue
			}
			errors = append(errors, fmt.Sprintf("mismatch at path %v: expected %v, got %v", d.Path, d.NewValues, d.OldValues))
		}
		if len(errors) > 0 {
			return fmt.Errorf(strings.Join(errors, "; "))
		}
		return nil
	}
}

// AssertOutputMatchesRegexp returns an assertion that checks if output matches a regexp
func AssertOutputMatchesRegexp(pattern string) func(*Task) error {
	return func(i *Task) error {
		actual := normalizeOutput(i.Actual.Output)
		matched, err := regexp.MatchString(pattern, actual)
		if err != nil {
			return fmt.Errorf("invalid regexp pattern %q: %v", pattern, err)
		}
		if !matched {
			return fmt.Errorf("output does not match pattern: %q", pattern)
		}
		return nil
	}
}

// Helper for robust JSON path skipping
func shouldSkipPath(path []jd.JsonNode, skipPaths []string) bool {
	var pathStrs []string
	for _, node := range path {
		pathStrs = append(pathStrs, node.Json())
	}
	joined := strings.Join(pathStrs, ".")
	for _, skip := range skipPaths {
		if joined == skip {
			return true
		}
	}
	return false
}
