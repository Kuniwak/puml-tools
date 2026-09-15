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
	flags := flag.NewFlagSet("pumlseqgroupcut", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		fmt.Fprintf(inout.Stderr, `Usage: pumlseqgroupcut -exact <label> [options] (<file>|< <file>)

Extracts the group block with the given label from a sequence diagram.
Only the declarations of participants referenced inside the extracted group
are kept; every other line (messages outside the group, unreferenced
participant declarations) is dropped. The group ... end frame itself is
kept as-is. If multiple groups match, all of them are processed.
The output @startuml does not carry over the input's diagram name; pass
-name to give it one.

Options:
`)
		flags.PrintDefaults()
		fmt.Fprintf(inout.Stderr, `
Examples:
  $ pumlseqgroupcut -exact Auth ./seq.puml
  @startuml
  participant B
  participant C
  group Auth
    B -> C: verify
    C -> B: ok
  end
  @enduml

  $ pumlseqgroupcut -exact Auth -name AuthExcerpt <./seq.puml
  @startuml AuthExcerpt
  ...
  @enduml
`)
	}

	var exact string
	flags.StringVar(&exact, "exact", "", "label of the group to extract (exact match against the text following the group keyword)")
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
