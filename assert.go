package iapetus

import (
	"errors"
	"fmt"
	"strings"

	jd "github.com/josephburnett/jd/lib"
)

func AssertByExitCode(i *Step) error {
	if i.Actual.ExitCode != i.Expected.ExitCode {
		return fmt.Errorf("exit code mismatch: expected %d, got %d", i.Expected.ExitCode, i.Actual.ExitCode)
	}
	return nil
}

func AssertByOutputString(i *Step) error {
	if i.Actual.Output != i.Expected.Output {
		return fmt.Errorf("output mismatch: expected %q, got %q", i.Expected.Output, i.Actual.Output)
	}
	return nil
}

func AssertByOutputJson(i *Step) error {
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

func AssertByContains(i *Step) error {
	for _, expected := range i.Expected.Contains {
		if !strings.Contains(i.Actual.Output, expected) {
			return fmt.Errorf("output does not contain expected substring: %q", expected)
		}
	}
	return nil
}

func AssertByError(i *Step) error {
	if i.Actual.Error != i.Expected.Error {
		return fmt.Errorf("error mismatch: expected %q, got %q", i.Expected.Error, i.Actual.Error)
	}
	return nil
}