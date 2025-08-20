package main

import (
	"fmt"
	"os"

	"github.com/yindia/iapetus"
)

func main() {
	ping := &iapetus.Task{
		Name:    "Ping Google",
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		Command: "ping",
		Args:    []string{"-c", "20", "baidu.com"},
		Asserts: []func(*iapetus.Task) error{
			iapetus.AssertExitCode(0),
			func(task *iapetus.Task) error {
				fmt.Println(task.Actual.Output)
				return nil
			},
		},
	}

	if err := ping.Run(); err != nil {
		fmt.Println("Error:", err)
	}
}
