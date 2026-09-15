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
	IDs           []string
	Name          string
	Reader        io.ReadCloser
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("pumlseqproj", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		fmt.Fprintf(inout.Stderr, `Usage: pumlseqproj -ids <a,b,...> [options] (<file>|< <file>)

Keeps only the messages of a sequence diagram whose sender or receiver is in the given
participant set, removing every other message line (this is the projection tr ↾ A of the
trace, the dual of pumlseqhide).
Notes attached to removed messages are removed along with them, and control blocks left empty
are pruned by the same rule as pumlseqhide. After projecting, any participant/actor
declaration that is no longer referenced is dropped (declarations named in -ids, and those
still referenced by a remaining message, note, ref over, or activate/deactivate, are kept).
A box that no longer encloses any declaration is also removed. ref over is never removed and
is left untouched (to project through its referenced content too, first run pumlseqexpand).
It is an error if -ids names a participant that is absent from the diagram (neither declared
nor referenced).
Pass -name to set the name on the output's @startuml (omit it to keep the input's name).
For PNG input, the embedded PlantUML metadata is decoded.

Options:
`)
		flags.PrintDefaults()
		fmt.Fprintf(inout.Stderr, `
Examples:
  $ pumlseqproj -ids B ./seq.puml
  $ pumlseqproj -ids B,C ./seq.puml
  $ pumlseqproj -ids B -name Focused <./seq.puml
`)
	}

	var ids string
	flags.StringVar(&ids, "ids", "", "participant ids to keep (comma separated, at least one)")
	var name string
	flags.StringVar(&name, "name", "", "name to set on the output's @startuml (omit to keep the input's name)")
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

	parsedIDs := pumlseq.SplitParticipants(ids)
	if len(parsedIDs) == 0 {
		return nil, errors.New("-ids is required (comma separated, at least one)")
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
		return &Options{CommonOptions: &commonOptions, IDs: parsedIDs, Name: name, Reader: f}, nil
	default:
		return &Options{CommonOptions: &commonOptions, IDs: parsedIDs, Name: name, Reader: inout.Stdin}, nil
	}
}
