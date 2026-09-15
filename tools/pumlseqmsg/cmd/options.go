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
	Common        []string
	Reader        io.ReadCloser
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("pumlseqmsg", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		fmt.Fprintf(inout.Stderr, `Usage: pumlseqmsg -common <a,b,...> [options] (<file>|< <file>)

Extracts the "common (synchronized) events" of a given participant set from a sequence
diagram - the messages whose sender and receiver are both in that set - and prints them as a
TSV of (sender<TAB>receiver<TAB>label).
The output is deduplicated as an event set (in order of first appearance). A singleton set
(e.g. -common Foo) yields Foo's self-transitions.
Use it to pull out the common events of a participant set you plan to fold with pumlseqfold,
and pipe the result into pumlseqhide --msg-file.
For PNG input, the embedded PlantUML metadata is decoded.

Options:
`)
		flags.PrintDefaults()
		fmt.Fprintf(inout.Stderr, `
Examples:
  $ pumlseqmsg -common B,C ./seq.puml
  B	C	fwd
  C	B	resp
`)
	}

	var common string
	flags.StringVar(&common, "common", "", "participant set to extract common events for (comma separated, at least one)")
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

	participants := pumlseq.SplitParticipants(common)
	if len(participants) == 0 {
		return nil, errors.New("-common is required (comma separated, at least one)")
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
		return &Options{CommonOptions: &commonOptions, Common: participants, Reader: f}, nil
	default:
		return &Options{CommonOptions: &commonOptions, Common: participants, Reader: inout.Stdin}, nil
	}
}
