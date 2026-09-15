package cmd

import (
	"fmt"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/version"
)

func MainCommandByArgs(args []string, inout *cli.ProcInout) int {
	options, err := ParseOptions(args, inout)
	if err != nil {
		fmt.Fprintln(inout.Stderr, err)
		return 1
	}

	if err := MainCommandByOptions(options, inout); err != nil {
		fmt.Fprintln(inout.Stderr, err)
		return 1
	}

	return 0
}

func MainCommandByOptions(options *Options, inout *cli.ProcInout) error {
	if options.CommonOptions.Help {
		return nil
	}

	if options.CommonOptions.Version {
		fmt.Fprintln(inout.Stdout, version.Version)
		return nil
	}

	if options.Short {
		for _, entry := range SelectEntries(Registry, options.ToolNames) {
			fmt.Fprintf(inout.Stdout, "%s: %s\n", entry.Name, entry.Summary)
		}
		return nil
	}

	child := &cli.ProcInout{
		Stdin:  inout.Stdin,
		Stdout: inout.Stdout,
		Stderr: inout.Stdout,
		Env:    inout.Env,
	}

	for _, entry := range SelectEntries(Registry, options.ToolNames) {
		fmt.Fprintf(inout.Stdout, "# %s\n\n```\n", entry.Name)
		if status := entry.Help([]string{"-h"}, child); status != 0 {
			return fmt.Errorf("cmd.MainCommandByOptions: %s: help exited with %d", entry.Name, status)
		}
		fmt.Fprint(inout.Stdout, "```\n\n")
	}
	return nil
}
