package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/pumlseqast"
	"github.com/Kuniwak/puml-tools/pumlseqtsv"
	"github.com/Kuniwak/puml-tools/tools"
)

type Options struct {
	CommonOptions  *tools.CommonOptions
	Phi            pumlseqast.Folding
	AbstractReader io.ReadCloser
	DetailReader   io.ReadCloser
	AbstractName   string
	DetailName     string
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("pumlseqrefinement", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		fmt.Fprintf(inout.Stderr, `Usage: pumlseqrefinement [options] <abstract> <detail>

Checks whether detail is a refinement of abstract (CSP trace refinement:
traces(abs_φ(detail)) ⊆ traces(abstract)). Use -map to give the folding map φ
that collapses detail's participants onto abstract's participants (any participant
not listed maps to itself by identity). abs_φ(detail) is detail folded through φ;
messages internal to a folded box remain visible as self-transitions (hide them
first with pumlseqhide if that's what you want).
Either abstract or detail may be read from stdin using -.

Semantics: this is a simplified, message-level model inspired by UML and CSP, not
the full interaction semantics of either. A whole message is one atomic event (UML
models its send and receive as two separate occurrences; this tool does not).
Juxtaposition is weak sequencing in the UML sense of ordering per lifeline: an
event on a lifeline cannot occur until everything written before it on that same
lifeline has finished, events on lifelines that share no participant have no
ordering between them, and wrapping them in a block does not change that. alt/opt
is choice; loop is zero or more repetitions; par is full interleaving;
group/critical/break are transparent (critical and break carry meanings in UML that
this model does not reproduce). Note that pumlseqsubset uses a different, flatter
model in which alt branches are concatenated rather than treated as a choice.

Note: the more messages that share no participant, the more the state count grows
exponentially (about 1,000 states for 10 pairwise-disjoint messages, about 260,000
for 18). In a diagram where pumlseqbreak has confirmed there is no break in
causality, consecutive messages always share a participant, so this matches a
total order and no blow-up occurs.

A "ref over" spans N lifelines to invoke another diagram and cannot be represented
as a single message, so an unexpanded ref is never silently ignored — it aborts.
Run the diagram through pumlseqexpand first.

Exit codes: 0 = refinement holds / 1 = refinement fails (the shortest
counterexample trace is printed to stderr) / 2 = usage or input error.

Options:
`)
		flags.PrintDefaults()
		fmt.Fprintf(inout.Stderr, `
Examples:
  $ pumlseqrefinement -map map.tsv abstract.puml detail.puml
  $ pumlseqhide --msg-file - detail.puml <internal.tsv | pumlseqrefinement abstract.puml -
`)
	}

	var mapFile string
	flags.StringVar(&mapFile, "map", "", "TSV folding map (detail ID<TAB>abstract ID); use - to read from stdin")
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
	if len(remaining) != 2 {
		return nil, errors.New("requires exactly two file paths: <abstract> <detail>")
	}
	abstractArg, detailArg := remaining[0], remaining[1]
	abstractStdin := abstractArg == "-"
	detailStdin := detailArg == "-"
	if abstractStdin && detailStdin {
		return nil, errors.New("cannot read both abstract and detail from stdin")
	}

	phi, err := readFolding(mapFile, inout, abstractStdin || detailStdin)
	if err != nil {
		return nil, err
	}

	abstractReader, err := cli.OpenInput(abstractArg, inout)
	if err != nil {
		return nil, err
	}
	detailReader, err := cli.OpenInput(detailArg, inout)
	if err != nil {
		if c, ok := abstractReader.(*os.File); ok {
			c.Close()
		}
		return nil, err
	}
	return &Options{
		CommonOptions:  &commonOptions,
		Phi:            phi,
		AbstractReader: abstractReader,
		DetailReader:   detailReader,
		AbstractName:   abstractArg,
		DetailName:     detailArg,
	}, nil
}

func readFolding(mapFile string, inout *cli.ProcInout, positionalUsesStdin bool) (pumlseqast.Folding, error) {
	if mapFile == "" {
		return pumlseqast.Folding{}, nil
	}
	if mapFile == "-" {
		if positionalUsesStdin {
			return nil, errors.New("cannot read both a diagram and -map from stdin")
		}
		return pumlseqtsv.ParseFolding(inout.Stdin)
	}
	f, err := os.Open(mapFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return pumlseqtsv.ParseFolding(f)
}
