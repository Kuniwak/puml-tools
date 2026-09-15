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
	Boundary      string
	Reader        io.ReadCloser
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("pumlc4bdyls", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		fmt.Fprintf(inout.Stderr, `Usage: pumlc4bdyls -boundary <id> (<file>|< <file>)

Lists, as a headered TSV, the elements contained (including nested) in the
C4 diagram boundary <id>. The columns are id / kind / external / boundary /
label, where boundary is each element's immediately enclosing boundary id.
Sub-boundary ids themselves are not listed, but the elements inside them
are. The output can be piped directly into pumlc4fold -ids-tsv (filter it
first with your favorite TSV tool if needed).
It is an error, and the command exits with a failure, if -boundary names a
boundary that doesn't exist in the diagram.
For a PNG input, the embedded PlantUML metadata is parsed.

Options:
`)
		flags.PrintDefaults()
		fmt.Fprintf(inout.Stderr, `
Examples:
  $ pumlc4bdyls -boundary shop ./c4.puml
  $ pumlc4bdyls -boundary shop <./c4.puml
`)
	}

	var boundary string
	flags.StringVar(&boundary, "boundary", "", "id of the boundary whose contained elements to list")
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

	if boundary == "" {
		return nil, errors.New("-boundary is required")
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
		return &Options{CommonOptions: &commonOptions, Boundary: boundary, Reader: f}, nil
	default:
		return &Options{CommonOptions: &commonOptions, Boundary: boundary, Reader: inout.Stdin}, nil
	}
}
