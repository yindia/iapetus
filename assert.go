package iapetus

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	jd "github.com/josephburnett/jd/lib"
)

func AssertByExitCode(i *Task) error {
	if i.Actual.ExitCode != i.Expected.ExitCode {
		return fmt.Errorf("exit code mismatch: expected %d, got %d", i.Expected.ExitCode, i.Actual.ExitCode)
	}
	return nil
}

func AssertByOutputString(i *Task) error {
	if i.Actual.Output != i.Expected.Output {
		return fmt.Errorf("output mismatch: expected %q, got %q", i.Expected.Output, i.Actual.Output)
	}
	return nil
}

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

func AssertByContains(i *Task) error {
	for _, expected := range i.Expected.Contains {
		if !strings.Contains(i.Actual.Output, expected) {
			return fmt.Errorf("output does not contain expected substring: %q", expected)
		}
	}
	return nil
}

func AssertByError(i *Task) error {
	if i.Actual.Error != i.Expected.Error {
		return fmt.Errorf("error mismatch: expected %q, got %q", i.Expected.Error, i.Actual.Error)
	}
	return nil
}

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
