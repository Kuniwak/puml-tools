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
	IDsTSVReader  io.ReadCloser
	NewID         string
	NewDesc       string
	Name          string
	Reader        io.ReadCloser
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("pumlc4fold", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		fmt.Fprintf(inout.Stderr, `Usage: pumlc4fold (-ids <a,b,...>|-ids-tsv <file>) -new-id <id> [options] (<file>|< <file>)

Folds several elements of a C4 diagram into one. The folded element is
always emitted as a System, so folding the containers of a Container
diagram (C4 Lv.2) back into their system yields a System Context diagram
(C4 Lv.1). Folding an arbitrary subset is allowed too; the output then
keeps whatever kinds were left untouched, and the !include is chosen to
cover the most detailed kind still present.
Every Rel to or from an element listed in -ids is rewritten to point at the
-new-id element instead. Rels between the folded elements (now self-Rels)
are dropped, and Rels that end up with the same (source, destination, label)
are merged into one. The output is a flat C4 diagram with no boundaries.
The elements to fold can be given directly with -ids (comma-separated), or
read from the id column of a headered TSV with -ids-tsv (intended for
piping in the output of pumlc4bdyls via process substitution; filter it
first with your favorite TSV tool if needed). -ids and -ids-tsv are
mutually exclusive, and exactly one is required. It is an error to name an
element that doesn't exist in the diagram, and an error if -new-id collides
with an existing element that isn't being folded. Pass -name to attach a
name to the output's @startuml (omit it to keep the input's name). For a
PNG input, the embedded PlantUML metadata is parsed.

Options:
`)
		flags.PrintDefaults()
		fmt.Fprintf(inout.Stderr, `
Examples:
  $ pumlc4fold -ids web,order_svc -new-id shop -new-desc "Shop" ./c4.puml
  $ pumlc4fold -ids-tsv <(pumlc4bdyls -boundary shop ./c4.puml) -new-id shop ./c4.puml
`)
	}

	var ids, idsTSV, newID, newDesc, name string
	flags.StringVar(&ids, "ids", "", "ids of the elements to fold (comma-separated, one or more)")
	flags.StringVar(&idsTSV, "ids-tsv", "", "path to a headered TSV whose id column lists the elements to fold (mutually exclusive with -ids)")
	flags.StringVar(&newID, "new-id", "", "id of the folded element")
	flags.StringVar(&newDesc, "new-desc", "", "display name of the folded element (optional; defaults to new-id)")
	flags.StringVar(&name, "name", "", "name to attach to the output's @startuml (defaults to the input's name)")
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

	if newID == "" {
		return nil, errors.New("-new-id is required")
	}

	if ids != "" && idsTSV != "" {
		return nil, errors.New("at most one of -ids/-ids-tsv can be set")
	}
	if ids == "" && idsTSV == "" {
		return nil, errors.New("one of -ids/-ids-tsv is required")
	}

	options := &Options{CommonOptions: &commonOptions, NewID: newID, NewDesc: newDesc, Name: name}
	if idsTSV != "" {
		f, err := os.Open(idsTSV)
		if err != nil {
			return nil, err
		}
		options.IDsTSVReader = f
	} else {
		parsedIDs := pumlseq.SplitParticipants(ids)
		if len(parsedIDs) == 0 {
			return nil, errors.New("-ids is empty (comma separated, at least one)")
		}
		options.IDs = parsedIDs
	}

	remaining := flags.Args()
	switch {
	case len(remaining) > 1:
		if options.IDsTSVReader != nil {
			options.IDsTSVReader.Close()
		}
		return nil, fmt.Errorf("too many arguments: accepts at most 1 file path")
	case len(remaining) == 1 && remaining[0] != "-":
		f, err := os.Open(remaining[0])
		if err != nil {
			if options.IDsTSVReader != nil {
				options.IDsTSVReader.Close()
			}
			return nil, err
		}
		options.Reader = f
	default:
		options.Reader = inout.Stdin
	}
	return options, nil
}
