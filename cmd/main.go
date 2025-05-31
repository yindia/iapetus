package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yindia/iapetus"
)

func printUsage() {
	fmt.Fprintf(os.Stderr, `iapetus: The open-source workflow engine for DevOps, CI/CD, and automation

Usage:
  iapetus run --config <workflow.yaml>

Options:
  --config   Path to workflow YAML config file (required)
  --help     Show this help message
`)
}

func main() {
	if len(os.Args) < 2 || os.Args[1] == "--help" || os.Args[1] == "-h" {
		printUsage()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "run":
		runCmd := flag.NewFlagSet("run", flag.ExitOnError)
		config := runCmd.String("config", "", "Path to workflow YAML config file (required)")
		runCmd.Usage = printUsage

		if err := runCmd.Parse(os.Args[2:]); err != nil {
			os.Exit(2)
		}
		if *config == "" {
			fmt.Fprintln(os.Stderr, "Error: --config is required")
			printUsage()
			os.Exit(2)
		}

		wf, err := iapetus.LoadWorkflowFromYAML(*config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load workflow: %v\n", err)
			os.Exit(1)
		}
		if err := wf.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Workflow failed: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(2)
	}
}
