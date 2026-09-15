package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/pumlseq"
	"github.com/Kuniwak/puml-tools/tools"
)

type Options struct {
	CommonOptions *tools.CommonOptions
	Matcher       pumlseq.MatchFunc
	Name          string
	Reader        io.ReadCloser
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("pumlseqgroupmask", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		fmt.Fprintf(inout.Stderr, `Usage: pumlseqgroupmask -exact <label> [options] (<file>|< <file>)

Removes the group block with the given label from a sequence diagram.
Declarations of participants that were referenced only inside the removed
group (participants no longer referenced anywhere in the remaining diagram)
are removed as well. Every other line is left as-is.
If multiple groups match, all of them are processed.
The output @startuml does not carry over the input's diagram name; pass
-name to give it one.

Options:
`)
		flags.PrintDefaults()
		fmt.Fprintf(inout.Stderr, `
Examples:
  $ pumlseqgroupmask -exact Auth ./seq.puml
  @startuml
  participant A
  participant B
  A -> B: login
  B -> A: done
  @enduml

  $ pumlseqgroupmask -exact Auth -name Overview <./seq.puml
  @startuml Overview
  ...
  @enduml
`)
	}

	var exact string
	flags.StringVar(&exact, "exact", "", "label of the group to remove (exact match against the text following the group keyword)")
	var name string
	flags.StringVar(&name, "name", "", "diagram name to give the output @startuml (default: no name)")
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

	matcher, err := selectMatcher(exact)
	if err != nil {
		return nil, err
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
		return &Options{CommonOptions: &commonOptions, Matcher: matcher, Name: name, Reader: f}, nil
	default:
		return &Options{CommonOptions: &commonOptions, Matcher: matcher, Name: name, Reader: inout.Stdin}, nil
	}
}

func selectMatcher(exact string) (pumlseq.MatchFunc, error) {
	if exact == "" {
		return nil, errors.New("-exact is required")
	}
	return pumlseq.Exact(exact), nil
}
