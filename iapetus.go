package iapetus

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/google/uuid"
)

type Step struct {
	Name          string
	Command       string
	Retries       int
	Args          []string
	Timeout       time.Duration
	Env           []string
	Expected      Output
	Actual        Output
	SkipJsonNodes []string
	Asserts       []func(*Step) error
}

type Output struct {
	ExitCode int
	Output   string
	Error    string
	Contains []string
}

func NewStep(name string, timeout time.Duration) *Step {
	if name == "" {
		name = "step-" + uuid.New().String()
	}
	return &Step{Name: name, Timeout: timeout}
}

func (s *Step) Run() error {
	if s.Name == "" {
		s.Name = "step-" + uuid.New().String()
	}
	if s.Retries == 0 {
		s.Retries = 1
	}
	log.Printf("Running step: %s", s.Name)
	cmd := exec.Command(s.Command, s.Args...)
	cmd.Env = append(os.Environ(), s.Env...)

	output, err := cmd.Output()
	s.Actual.ExitCode = getExitCode(err)
	s.Actual.Output = string(output)

	if err != nil {
		fmt.Println(s.Actual.Error)
		s.Actual.Error = err.Error()
	}

	for _, assert := range s.Asserts {
		if err := assert(s); err != nil {
			log.Printf("Assertion failed: %s", err.Error())
			return err
		}
	}
	return nil
}

func (s *Step) AddAssertion(assert func(*Step) error) *Step {
	s.Asserts = append(s.Asserts, assert)
	return s
}

func (s *Step) AddContains(contains ...string) *Step {
	s.Expected.Contains = append(s.Expected.Contains, contains...)
	return s
}

func (s *Step) AddEnv(env ...string) *Step {
	s.Env = append(s.Env, env...)
	return s
}

func (s *Step) AddArgs(args ...string) *Step {
	s.Args = append(s.Args, args...)
	return s
}

func (s *Step) AddSkipJsonNodes(skipJsonNodes ...string) *Step {
	s.SkipJsonNodes = append(s.SkipJsonNodes, skipJsonNodes...)
	return s
}

func (s *Step) AddExpected(expected Output) *Step {
	s.Expected = expected
	return s
}

func (s *Step) AddCommand(command string) *Step {
	s.Command = command
	return s
}
