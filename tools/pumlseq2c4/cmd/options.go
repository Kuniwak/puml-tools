package cmd

import (
	"errors"
	"flag"
	"fmt"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/tools"
)

const (
	BoundaryNone = "none"
	BoundaryBox  = "box"
)

type Options struct {
	CommonOptions *tools.CommonOptions
	Name          string
	Boundary      string
	SeqInputs     []tools.NamedInput
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("pumlseq2c4", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		fmt.Fprintf(inout.Stderr, `Usage: pumlseq2c4 [options] (<file>|-)...

Reads one or more PlantUML sequence diagrams and builds a single
C4-PlantUML diagram from their union. Each participant becomes a C4
element (actor becomes Person, everything else becomes System), and each
message becomes a Rel (with the message body as its label). With
multiple inputs, elements are deduplicated by id and Rels by (sender,
receiver, label) across inputs, and self-messages are dropped. No
diagram name is set (pass -name to set one). Aggregating or renaming
participants is out of scope; inputs are assumed to already correspond
1:1 with C4 (use pumlseqfold beforehand if you need to merge them). A
ref over block does not contribute to C4 (it's a call into another
diagram, so it produces neither elements nor Rels); when one is found, a
warning is printed to stderr. Either pass the referenced diagram as an
additional positional argument, or run the input through pumlseqexpand.
Passing -boundary box turns each sequence diagram's box into a C4
System_Boundary, placing the participants declared inside that box into
the boundary. Participants that don't belong to a box stay outside any
boundary. With multiple inputs, it is an error if the same participant
belongs to a different box in different diagrams.
Omit the arguments to read a single input from stdin. Either argument may
be - for stdin. For a PNG input, the embedded PlantUML metadata is
parsed.

Options:
`)
		flags.PrintDefaults()
		fmt.Fprintf(inout.Stderr, `
Examples:
  $ pumlseq2c4 ./seq-01.puml ./seq-02.puml
  @startuml
  !include <C4/C4_Context>

  System(a, "A")
  System(b, "B")

  Rel(a, b, "fwd")
  @enduml
`)
	}

	var name string
	flags.StringVar(&name, "name", "", "name of the output C4 diagram (attached to @startuml); no name if omitted")
	var boundary string
	flags.StringVar(&boundary, "boundary", BoundaryNone, fmt.Sprintf("how to build boundaries: %s emits no boundaries (default), %s turns each sequence diagram's box into a System_Boundary", BoundaryNone, BoundaryBox))
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
	if boundary != BoundaryNone && boundary != BoundaryBox {
		return nil, fmt.Errorf("unknown -boundary value: %q (must be one of: %s, %s)", boundary, BoundaryNone, BoundaryBox)
	}
	if commonOptions.Version {
		return &Options{CommonOptions: &commonOptions}, nil
	}

	remaining := flags.Args()
	if len(remaining) == 0 {
		return &Options{CommonOptions: &commonOptions, Name: name, Boundary: boundary, SeqInputs: []tools.NamedInput{{Name: "-", Reader: inout.Stdin}}}, nil
	}

	dashes := 0
	for _, a := range remaining {
		if a == "-" {
			dashes++
		}
	}
	if dashes > 1 {
		return nil, errors.New("at most one argument can be - (stdin)")
	}

	var seqInputs []tools.NamedInput
	for _, a := range remaining {
		r, err := cli.OpenInput(a, inout)
		if err != nil {
			for _, opened := range seqInputs {
				opened.Reader.Close()
			}
			return nil, err
		}
		seqInputs = append(seqInputs, tools.NamedInput{Name: a, Reader: r})
	}
	return &Options{CommonOptions: &commonOptions, Name: name, Boundary: boundary, SeqInputs: seqInputs}, nil
}
