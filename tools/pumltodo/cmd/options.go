package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/tools"
)

type Options struct {
	CommonOptions *tools.CommonOptions
	Reader        io.ReadCloser
	Name          string
	Pattern       *regexp.Regexp
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("pumltodo", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		fmt.Fprintf(inout.Stderr, `Usage: pumltodo [options] (<file>|< <file>)

Extracts TODOs matching --regexp from PlantUML line comments ('), block comments
(/' '/), note/hnote/rnote, and ref blocks, and lists them as 3-column TSV: file
name, line number, content.
For a multi-line block, a single matching line counts the whole block as one hit.
If the input is a PNG, the PlantUML embedded in its metadata is parsed instead.

Options:
`)
		flags.PrintDefaults()
		fmt.Fprintf(inout.Stderr, `
Examples:
  $ pumltodo ./diagram.puml
  ./diagram.puml	3	TODO: ...

  $ pumltodo --regexp '(TODO|FIXME)' <./diagram.puml
  -	3	TODO: ...
`)
	}

	pattern := flags.String("regexp", "(TODO|todo)", "regular expression to match")

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

	re, err := regexp.Compile(*pattern)
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
		return &Options{CommonOptions: &commonOptions, Reader: f, Name: remaining[0], Pattern: re}, nil
	default:
		return &Options{CommonOptions: &commonOptions, Reader: inout.Stdin, Name: "-", Pattern: re}, nil
	}
}
