package cmd

import (
	"fmt"
	"io"

	"github.com/Kuniwak/puml-tools/c4ast"
	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/pumlseqast"
	"github.com/Kuniwak/puml-tools/pumlseqc4"
	"github.com/Kuniwak/puml-tools/tools/pumlparse/pumlpng"
	"github.com/Kuniwak/puml-tools/version"
)

func MainCommandByArgs(args []string, inout *cli.ProcInout) int {
	options, err := ParseOptions(args, inout)
	if err != nil {
		fmt.Fprintln(inout.Stderr, err)
		return 1
	}

	if err := MainCommandByOptions(options, inout); err != nil {
		fmt.Fprintln(inout.Stderr, err)
		return 1
	}

	return 0
}

func MainCommandByOptions(options *Options, inout *cli.ProcInout) error {
	if options.CommonOptions.Help {
		return nil
	}

	if options.CommonOptions.Version {
		fmt.Fprintln(inout.Stdout, version.Version)
		return nil
	}

	ds := make([]*pumlseqast.Diagram, 0, len(options.SeqInputs))
	for _, in := range options.SeqInputs {
		d, err := readDiagram(in.Reader)
		if err != nil {
			return fmt.Errorf("%s: %w", in.Name, err)
		}
		for _, w := range pumlseqc4.RefWarnings(d) {
			fmt.Fprintf(inout.Stderr, "%s:%d: warning: %s\n", in.Name, w.Line, w.Reason())
		}
		ds = append(ds, d)
	}

	c4d, err := toC4(ds, options.Boundary)
	if err != nil {
		return err
	}
	if options.Name != "" {
		c4d.Name = options.Name
	}

	_, err = io.WriteString(inout.Stdout, c4d.String())
	return err
}

func toC4(ds []*pumlseqast.Diagram, boundary string) (*c4ast.Diagram, error) {
	if boundary == BoundaryBox {
		return pumlseqc4.FromSeqsWithBoxBoundaries(ds)
	}
	return pumlseqc4.FromSeqs(ds), nil
}

func readDiagram(r io.ReadCloser) (*pumlseqast.Diagram, error) {
	defer r.Close()
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	src, err := pumlpng.Source(b)
	if err != nil {
		return nil, err
	}

	return pumlseqast.Parse(string(src))
}
