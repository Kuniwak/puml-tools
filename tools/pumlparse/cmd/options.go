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
	Raw           bool
	Reader        io.ReadCloser
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("pumlparse", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		fmt.Fprintf(inout.Stderr, `Usage: pumlparse [options] (<file>|< <file>)

Extracts PlantUML source embedded in a PNG's metadata, parses it into an AST, and
renders it back to text (a round trip). If the input file is already plain text,
it is parsed and rendered back the same way.
With --raw, the extracted PlantUML is printed as-is, without parsing into an AST.

Options:
`)
		flags.PrintDefaults()
		fmt.Fprintf(inout.Stderr, `
Examples:
  $ pumlparse <./input.puml.png
  @startuml
  ...
  @enduml

  $ pumlparse ./input.puml.png
  @startuml
  ...
  @enduml
`)
	}
	var raw bool
	flags.BoolVar(&raw, "raw", false, "print the PlantUML embedded in the metadata as-is, without parsing it into an AST")
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
		return &Options{CommonOptions: &commonOptions, Raw: raw, Reader: f}, nil
	default:
		return &Options{CommonOptions: &commonOptions, Raw: raw, Reader: inout.Stdin}, nil
	}
}
