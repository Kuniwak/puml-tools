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
	NewID         string
	NewDesc       string
	Name          string
	Reader        io.ReadCloser
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("pumlseqfold", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		fmt.Fprintf(inout.Stderr, `Usage: pumlseqfold -ids <a,b,...> -new-id <id> [options] (<file>|< <file>)

Folds several participants of a sequence diagram into one (summarization). Every message and
reference (activate/deactivate, note over, ref over) addressed to or from a participant in
-ids is rewritten to belong to the -new-id participant. If folding collapses a ref over's
targets into one, duplicates are removed.
Messages between folded participants remain as self-transitions (remove them with pumlseqhide).
It is an error if -ids names a participant that does not exist in the diagram, and likewise if
-new-id collides with an existing participant outside the fold.
Control blocks left empty by folding or a prior hide (empty alt/par branches, or empty
opt/loop/group/break, etc.) are pruned automatically, and a par/alt reduced to a single branch
has its contents hoisted to the parent level (recursively, from the innermost block out, and
idempotently).
Pass -name to set the name on the output's @startuml (omit it to keep the input's name).
For PNG input, the embedded PlantUML metadata is decoded.
Note: even when the input uses CRLF line endings, newly inserted participant declaration lines
end with LF (the mixed line endings do not affect how PlantUML interprets the diagram).

Options:
`)
		flags.PrintDefaults()
		fmt.Fprintf(inout.Stderr, `
Examples:
  $ pumlseqfold -ids B,C -new-id BE -new-desc "Backend" ./seq.puml
  $ pumlseqfold -ids B,C -new-id BE -name Folded <./seq.puml
`)
	}

	var ids string
	flags.StringVar(&ids, "ids", "", "participant ids to fold (comma separated, at least one)")
	var newID string
	flags.StringVar(&newID, "new-id", "", "id of the resulting folded participant")
	var newDesc string
	flags.StringVar(&newDesc, "new-desc", "", "description for the resulting folded participant (optional; omit for no description)")
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
	if newID == "" {
		return nil, errors.New("-new-id is required")
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
		return &Options{CommonOptions: &commonOptions, IDs: parsedIDs, NewID: newID, NewDesc: newDesc, Name: name, Reader: f}, nil
	default:
		return &Options{CommonOptions: &commonOptions, IDs: parsedIDs, NewID: newID, NewDesc: newDesc, Name: name, Reader: inout.Stdin}, nil
	}
}
