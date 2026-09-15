package cmd

import (
	"errors"
	"fmt"
	"io"

	"github.com/Kuniwak/puml-tools/c4ast"
	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/pumlseqast"
	"github.com/Kuniwak/puml-tools/pumlseqc4"
	"github.com/Kuniwak/puml-tools/tools"
	"github.com/Kuniwak/puml-tools/tools/pumlparse/pumlpng"
	"github.com/Kuniwak/puml-tools/version"
)

func MainCommandByArgs(args []string, inout *cli.ProcInout) int {
	options, err := ParseOptions(args, inout)
	if err != nil {
		fmt.Fprintln(inout.Stderr, err)
		return 2
	}

	code, err := MainCommandByOptions(options, inout)
	if err != nil {
		fmt.Fprintln(inout.Stderr, err)
		return 2
	}

	return code
}

func MainCommandByOptions(options *Options, inout *cli.ProcInout) (int, error) {
	if options.CommonOptions.Help {
		return 0, nil
	}

	if options.CommonOptions.Version {
		fmt.Fprintln(inout.Stdout, version.Version)
		return 0, nil
	}

	defer func() {
		for _, in := range options.SeqInputs {
			in.Reader.Close()
		}
	}()
	defer options.C4Reader.Close()

	seqDs := make([]*pumlseqast.Diagram, 0, len(options.SeqInputs))
	for _, in := range options.SeqInputs {
		d, err := readSeq(in.Reader)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", in.Name, err)
		}
		for _, w := range pumlseqc4.RefWarnings(d) {
			fmt.Fprintf(inout.Stderr, "%s:%d: warning: %s\n", in.Name, w.Line, w.Reason())
		}
		seqDs = append(seqDs, d)
	}
	c4D, err := readC4(options.C4Reader)
	if err != nil {
		return 0, err
	}

	diff, err := pumlseqc4.CompareSeqs(seqDs, c4D)
	if err != nil {
		return 0, namedError(err, options.SeqInputs)
	}
	switch options.Method {
	case MethodSubset:
		if diff.IsSubset() {
			fmt.Fprintln(inout.Stdout, "OK: c4 ⊆ seq")
			return 0, nil
		}
		fmt.Fprint(inout.Stdout, diff.ReportSubset())
		n := len(diff.C4OnlyNodes) + len(diff.C4OnlyEdges)
		fmt.Fprintf(inout.Stderr, "c4 is not a subset of seq: %d items\n", n)
		return 1, nil
	case MethodSuperset:
		if diff.IsSuperset() {
			fmt.Fprintln(inout.Stdout, "OK: seq \u2286 c4")
			return 0, nil
		}
		fmt.Fprint(inout.Stdout, diff.ReportSuperset())
		n := len(diff.SeqOnlyNodes) + len(diff.SeqOnlyEdges)
		fmt.Fprintf(inout.Stderr, "c4 is not a superset of seq: %d items\n", n)
		return 1, nil
	default:
		if diff.IsEqual() {
			fmt.Fprintln(inout.Stdout, "OK: consistent")
			return 0, nil
		}
		fmt.Fprint(inout.Stdout, diff.ReportEq())
		n := len(diff.SeqOnlyNodes) + len(diff.C4OnlyNodes) + len(diff.SeqOnlyEdges) + len(diff.C4OnlyEdges)
		fmt.Fprintf(inout.Stderr, "inconsistent: %d differences\n", n)
		return 1, nil
	}
}

func namedError(err error, seqInputs []tools.NamedInput) error {
	var conflict *pumlseqc4.ParticipantConflict
	if !errors.As(err, &conflict) {
		return err
	}
	return fmt.Errorf("participant %q conflicts between %s and %s: %s %q vs %s %q",
		conflict.ID,
		seqInputs[conflict.FirstIndex].Name, seqInputs[conflict.Index].Name,
		conflict.FirstKind, conflict.FirstLabel, conflict.Kind, conflict.Label)
}

func readSeq(r io.Reader) (*pumlseqast.Diagram, error) {
	src, err := readSource(r)
	if err != nil {
		return nil, err
	}
	return pumlseqast.Parse(string(src))
}

func readC4(r io.Reader) (*c4ast.Diagram, error) {
	src, err := readSource(r)
	if err != nil {
		return nil, err
	}
	return c4ast.Parse(string(src))
}

func readSource(r io.Reader) ([]byte, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return pumlpng.Source(b)
}
