package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/outfiles"
	"github.com/Kuniwak/puml-tools/tools"
)

type Options struct {
	CommonOptions *tools.CommonOptions
	OutDir        string
	Basename      string
	Reader        io.ReadCloser
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("pumlseqgroupsplit", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		fmt.Fprintf(inout.Stderr, `Usage: pumlseqgroupsplit -o <out_dir> [options] (<file>|< <file>)

Splits a sequence diagram whose every message is enclosed in a group into
one *.puml file per group. The unit of the split is the outermost group;
a nested group stays inside its enclosing group's file. Each file is cut
out the same way as pumlseqgroupcut (only the declarations of participants
referenced inside that group and the group ... end frame are kept, every
other line is dropped). The output @startuml is left unnamed.
File names follow <out_dir>/<basename>_<idx>_<group title>.puml, where idx
starts at 1 and is zero-padded to the number of digits in the group count.
Spaces in the title and characters not allowed in file names (/ \ : * ? "
< > |) are replaced with "_". A group with an empty title produces
<basename>_<idx>.puml.
basename defaults to the input file name with a trailing .png or .puml
stripped. When reading from stdin, -basename is required, and a -basename
containing a path separator (/ \) is an error.
<out_dir> is created if it does not exist, and files with the same name
are overwritten. Every file is written to a temporary file first and only
then renamed into place, so a failure while generating or writing content
leaves <out_dir> untouched. Once the renames begin they are not rolled
back: if a rename itself fails, the files already renamed stay in place.
Written paths are printed to stdout, one per line.
It is an error if there are no groups at all, or if even one message is
not enclosed in any group (such a message would silently fall out of
every file). ref over outside a group is not checked; if its target is
also enclosed in a group, run pumlseqexpand first so that the expanded
messages end up inside the group before splitting.
For PNG input, the embedded PlantUML metadata is read.

Options:
`)
		flags.PrintDefaults()
		fmt.Fprintf(inout.Stderr, `
Examples:
  $ pumlseqgroupsplit -o ./out ./seq.puml
  ./out/seq_1_Login.puml
  ./out/seq_2_Order.puml

  $ pumlseqgroupsplit -o ./out -basename flow <./seq.puml
  ./out/flow_1_Login.puml
  ./out/flow_2_Order.puml
`)
	}

	var outDir string
	flags.StringVar(&outDir, "o", "", "directory to write the split *.puml files into (created if it does not exist)")
	var basename string
	flags.StringVar(&basename, "basename", "", "base name for output files (default: derived from the input file name; required when reading from stdin)")
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

	if outDir == "" {
		return nil, errors.New("-o is required")
	}

	remaining := flags.Args()
	if len(remaining) > 1 {
		return nil, fmt.Errorf("too many arguments: accepts at most 1 file path")
	}

	filePath := ""
	if len(remaining) == 1 && remaining[0] != "-" {
		filePath = remaining[0]
	}
	resolved, err := outfiles.ResolveBasename(basename, filePath)
	if err != nil {
		if errors.Is(err, outfiles.ErrBasenameEmpty) {
			return nil, errors.New("-basename is required when reading the diagram from stdin")
		}
		return nil, fmt.Errorf("-basename %q: %w", basename, err)
	}

	reader, err := cli.OpenInput(inputName(filePath), inout)
	if err != nil {
		return nil, err
	}
	return &Options{CommonOptions: &commonOptions, OutDir: outDir, Basename: resolved, Reader: reader}, nil
}

func inputName(filePath string) string {
	if filePath == "" {
		return "-"
	}
	return filePath
}
