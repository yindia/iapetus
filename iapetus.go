package iapetus

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/google/uuid"
)

type Task struct {
	Name          string
	Command       string
	Retries       int
	Args          []string
	Timeout       time.Duration
	Env           []string
	Expected      Output
	Actual        Output
	SkipJsonNodes []string
	Asserts       []func(*Task) error
}

type Output struct {
	ExitCode int
	Output   string
	Error    string
	Contains []string
}

func NewTask(name string, timeout time.Duration) *Task {
	if name == "" {
		name = "task-" + uuid.New().String()
	}
	return &Task{Name: name, Timeout: timeout}
}

func (s *Task) Run() error {
	if s.Name == "" {
		s.Name = "task-" + uuid.New().String()
	}
	if s.Retries == 0 {
		s.Retries = 1
	}
	log.Printf("Running task: %s", s.Name)

	for attempt := 1; attempt <= s.Retries; attempt++ {
		log.Printf("Attempt %d of %d for task: %s", attempt, s.Retries, s.Name)

		cmd := exec.Command(s.Command, s.Args...)
		cmd.Env = append(os.Environ(), s.Env...)

		output, err := cmd.CombinedOutput()
		s.Actual.ExitCode = getExitCode(err)
		s.Actual.Output = string(output)

		if err != nil {
			s.Actual.Error = err.Error()
			log.Printf("Error executing task %s: %v", s.Name, err)
		}

		for _, assert := range s.Asserts {
			if err := assert(s); err != nil {
				log.Printf("Assertion failed for task %s: %v", s.Name, err)
				if attempt < s.Retries {
					log.Printf("Retrying task %s after failure", s.Name)
					time.Sleep(1 * time.Second)
					continue
				}
				return fmt.Errorf("assertion failed for task %s: %w", s.Name, err)
			}
		}
		break
	}

	return nil
}

func (s *Task) AddAssertion(assert func(*Task) error) *Task {
	s.Asserts = append(s.Asserts, assert)
	return s
}

func (s *Task) AddContains(contains ...string) *Task {
	s.Expected.Contains = append(s.Expected.Contains, contains...)
	return s
}

func (s *Task) AddEnv(env ...string) *Task {
	s.Env = append(s.Env, env...)
	return s
}

func (s *Task) AddArgs(args ...string) *Task {
	s.Args = append(s.Args, args...)
	return s
}

func (s *Task) AddSkipJsonNodes(skipJsonNodes ...string) *Task {
	s.SkipJsonNodes = append(s.SkipJsonNodes, skipJsonNodes...)
	return s
}

func (s *Task) AddExpected(expected Output) *Task {
	s.Expected = expected
	return s
}

func (s *Task) AddCommand(command string) *Task {
	s.Command = command
	return s
}
