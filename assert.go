package iapetus

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	jd "github.com/josephburnett/jd/lib"
)

// AssertByExitCode verifies that the actual exit code matches the expected exit code.
// Returns an error if there's a mismatch, nil otherwise.
func AssertByExitCode(i *Task) error {
	if i.Actual.ExitCode != i.Expected.ExitCode {
		return fmt.Errorf("exit code mismatch: expected %d, got %d", i.Expected.ExitCode, i.Actual.ExitCode)
	}
	return nil
}

// AssertByOutputString compares the actual output string with the expected output string.
// Returns an error if there's a mismatch, nil otherwise.
func AssertByOutputString(i *Task) error {
	if i.Actual.Output != i.Expected.Output {
		return fmt.Errorf("output mismatch: expected %q, got %q", i.Expected.Output, i.Actual.Output)
	}
	return nil
}

// AssertByOutputJson compares JSON outputs by parsing both expected and actual outputs.
// It normalizes line breaks and supports skipping specific JSON nodes during comparison.
// Returns an error if there's a parsing error or mismatch, nil otherwise.
func AssertByOutputJson(i *Task) error {
	expectation, err := jd.ReadJsonString(i.Expected.Output)
	if err != nil {
		return errors.New("Failed to read expectation: " + err.Error())
	}

	// normalize linebreaks
	parsedOutput, err := jd.ReadJsonString(strings.ReplaceAll(i.Actual.Output, "\\r\\n", "\\n"))
	if err != nil {
		return errors.New("Failed to parse output: " + err.Error())
	}

	diff := expectation.Diff(parsedOutput)
	if len(diff) != 0 {
		var path jd.JsonNode
		for _, d := range diff {
			path = d.Path[len(d.Path)-1]
			for _, skip := range i.SkipJsonNodes {
				if path.Json() == skip {
					continue
				}
				return fmt.Errorf(
					"mismatch at path %v. Expected json: %v, but found: %v",
					d.Path, d.NewValues, d.OldValues,
				)
			}

		}
	}

	return nil
}

// AssertByContains checks if the actual output contains all expected substrings.
// Returns an error if any substring is missing, nil otherwise.
func AssertByContains(i *Task) error {
	for _, expected := range i.Expected.Contains {
		if !strings.Contains(i.Actual.Output, expected) {
			return fmt.Errorf("output does not contain expected substring: %q", expected)
		}
	}
	return nil
}

// AssertByError verifies that the actual error matches the expected error.
// Returns an error if there's a mismatch, nil otherwise.
func AssertByError(i *Task) error {
	if i.Actual.Error != i.Expected.Error {
		return fmt.Errorf("error mismatch: expected %q, got %q", i.Expected.Error, i.Actual.Error)
	}
	return nil
}

// AssertByRegexp checks if the actual output matches all expected regular expression patterns.
// Returns an error if any pattern doesn't match or is invalid, nil otherwise.
func AssertByRegexp(i *Task) error {
	for _, pattern := range i.Expected.Patterns {
		matched, err := regexp.MatchString(pattern, i.Actual.Output)
		if err != nil {
			return fmt.Errorf("invalid regexp pattern %q: %v", pattern, err)
		}
		if !matched {
			return fmt.Errorf("output does not match pattern: %q", pattern)
		}
	}
	return nil
}
