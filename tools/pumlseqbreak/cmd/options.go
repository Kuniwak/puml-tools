package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/tools"
)

type Options struct {
	CommonOptions *tools.CommonOptions
	Reader        io.ReadCloser
	Name          string
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("pumlseqbreak", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		fmt.Fprintf(inout.Stderr, `Usage: pumlseqbreak [options] (<file>|< <file>)

Scans a sequence diagram's message arrows and reports each message whose
sender has no preceding trigger (a break in the causal chain), listing
them as a 3-column TSV of file name, line number, and reason.
Each message is checked only against the immediately preceding message
(a chain a->b/b->c or a broadcast a->b/a->c is fine).
A "..." line resets causality and always permits the next message.
A ref over is a call into another diagram, so it is treated as a single
opaque event involving its target lifelines (after the ref, a send from
any of its targets counts as a trigger; a ref whose targets have no
trigger at all is itself reported as a break).
Blocks such as alt/opt/loop/group/par are checked against the context in
effect at entry, once per branch.
Exit status is 1 if one or more breaks are found. For PNG input, the
embedded PlantUML metadata is read.

Options:
`)
		flags.PrintDefaults()
		fmt.Fprintf(inout.Stderr, `
Examples:
  $ pumlseqbreak ./diagram.puml
  ./diagram.puml	3	sender c has no preceding trigger
`)
	}

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

	remaining := flags.Args()
	switch {
	case len(remaining) > 1:
		return nil, fmt.Errorf("too many arguments: accepts at most 1 file path")
	case len(remaining) == 1 && remaining[0] != "-":
		f, err := os.Open(remaining[0])
		if err != nil {
			return nil, err
		}
		return &Options{CommonOptions: &commonOptions, Reader: f, Name: remaining[0]}, nil
	default:
		return &Options{CommonOptions: &commonOptions, Reader: inout.Stdin, Name: "-"}, nil
	}
}
