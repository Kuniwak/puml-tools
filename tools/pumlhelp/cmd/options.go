package cmd

import (
	"errors"
	"flag"
	"fmt"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/tools"
)

type Options struct {
	CommonOptions *tools.CommonOptions
	Short         bool
	ToolNames     []string
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("pumlhelp", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		fmt.Fprintf(inout.Stderr, `Usage: pumlhelp [options] [tool ...]

Prints the help of every PlantUML tool, gathered together as Markdown.
With no arguments it shows every tool; naming one or more tools shows only those.
Tools are always listed in dictionary order by name, not in the order given.

Options:
`)
		flags.PrintDefaults()
		fmt.Fprintf(inout.Stderr, `
Examples:
  $ pumlhelp
  $ pumlhelp -short
  $ pumlhelp pumlparse
`)
	}
	var short bool
	flags.BoolVar(&short, "short", false, "print only a one-line summary of each tool")
	commonRawOptions := &tools.CommonRawOptions{}
	tools.DeclareCommonFlags(flags, commonRawOptions)

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return &Options{CommonOptions: &tools.CommonOptions{Help: true}}, nil
		}
		return nil, err
	}

	commonOptions, err := tools.ValidateCommonOptions(commonRawOptions)
	if err != nil {
		return nil, err
	}

	if commonOptions.Version {
		return &Options{CommonOptions: &commonOptions}, nil
	}

	names := flags.Args()
	for _, name := range names {
		if !KnownTool(Registry, name) {
			return nil, fmt.Errorf("cmd.ParseOptions: unknown tool: %s", name)
		}
	}

	return &Options{CommonOptions: &commonOptions, Short: short, ToolNames: names}, nil
}
