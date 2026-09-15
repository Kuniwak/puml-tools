package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/tools"
)

const (
	MethodEq       = "eq"
	MethodSubset   = "subset"
	MethodSuperset = "superset"
)

type Options struct {
	CommonOptions *tools.CommonOptions
	Method        string
	SeqInputs     []tools.NamedInput
	C4Reader      io.ReadCloser
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("pumlc42seqchk", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		fmt.Fprintf(inout.Stderr, `Usage: pumlc42seqchk (-seq <path|->|<path|->...) -c4 <path|-> [-method eq|subset|superset]

Checks whether a sequence diagram and a C4 diagram are consistent with each
other. Participants and C4 elements are compared by id, and messages and
Rels are compared as directed (sender, receiver) pairs (labels are ignored,
self-relations are dropped, and boundaries don't count as elements).
-method eq (the default) requires the node sets and directed edge sets to
match exactly; -method subset requires C4 to be a subset of the sequence
diagram (C4 subset-of SEQ); -method superset requires C4 to be a superset
of the sequence diagram (SEQ subset-of C4).
Multiple sequence diagrams can be given as positional arguments. In that
case SEQ is their union (participants deduplicated by id, messages
deduplicated as directed pairs), so you can check whether several sequence
diagrams combined (eq) match a single C4 diagram. It is an error if the
same participant id is declared with a different display name or kind in
different diagrams, since they can't then be merged.
A ref over block does not contribute to C4 (it's a call into another
diagram, so it produces neither elements nor Rels); when one is found, a
warning is printed to stderr. The check still proceeds, but it may report
"inconsistent" purely because the referenced relationships are missing, so
either pass the referenced diagram as an additional positional argument, or
run the input through pumlseqexpand.
Returns exit code 0 if consistent, 1 if inconsistent, and 2 on error. When
inconsistent, the differences are listed on stdout. At most one input may
be - to read from stdin.
For a PNG input, the embedded PlantUML metadata is parsed.

Options:
`)
		flags.PrintDefaults()
		fmt.Fprintf(inout.Stderr, `
Examples:
  $ pumlc42seqchk -method subset -seq seq.puml -c4 c4.puml; echo $?
  elements only in C4 (not in SEQ): 1
    - foo ("Foo")
  1

  $ pumlc42seqchk -method eq -c4 c4.puml seq-01.puml seq-02.puml; echo $?
  OK: consistent
  0
`)
	}

	var seqPath, c4Path, method string
	flags.StringVar(&seqPath, "seq", "", "path to the sequence diagram (- for stdin); mutually exclusive with positional arguments")
	flags.StringVar(&c4Path, "c4", "", "path to the C4 diagram (- for stdin)")
	flags.StringVar(&method, "method", MethodEq, "check method eq|subset|superset (default eq)")
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

	if method != MethodEq && method != MethodSubset && method != MethodSuperset {
		return nil, fmt.Errorf("invalid -method %q: must be %q, %q or %q", method, MethodEq, MethodSubset, MethodSuperset)
	}
	remaining := flags.Args()
	if seqPath != "" && len(remaining) > 0 {
		return nil, errors.New("-seq and positional sequence diagram arguments are mutually exclusive")
	}
	seqPaths := remaining
	if seqPath != "" {
		seqPaths = []string{seqPath}
	}
	if len(seqPaths) == 0 || c4Path == "" {
		return nil, errors.New("-seq (or positional sequence diagram arguments) and -c4 are both required")
	}

	dashes := 0
	for _, p := range append(append([]string{}, seqPaths...), c4Path) {
		if p == "-" {
			dashes++
		}
	}
	if dashes > 1 {
		return nil, errors.New("at most one input can be - (stdin)")
	}

	var seqInputs []tools.NamedInput
	closeSeqInputs := func() {
		for _, in := range seqInputs {
			in.Reader.Close()
		}
	}
	for _, p := range seqPaths {
		r, err := cli.OpenInput(p, inout)
		if err != nil {
			closeSeqInputs()
			return nil, err
		}
		seqInputs = append(seqInputs, tools.NamedInput{Name: p, Reader: r})
	}
	c4Reader, err := cli.OpenInput(c4Path, inout)
	if err != nil {
		closeSeqInputs()
		return nil, err
	}

	return &Options{
		CommonOptions: &commonOptions,
		Method:        method,
		SeqInputs:     seqInputs,
		C4Reader:      c4Reader,
	}, nil
}
