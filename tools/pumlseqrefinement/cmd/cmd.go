package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/csp"
	"github.com/Kuniwak/puml-tools/pumlseqast"
	"github.com/Kuniwak/puml-tools/tools/pumlparse/pumlpng"
	"github.com/Kuniwak/puml-tools/version"
)

const (
	exitRefines    = 0
	exitNotRefines = 1
	exitError      = 2
)

func MainCommandByArgs(args []string, inout *cli.ProcInout) int {
	options, err := ParseOptions(args, inout)
	if err != nil {
		fmt.Fprintln(inout.Stderr, err)
		return exitError
	}
	return MainCommandByOptions(options, inout)
}

func MainCommandByOptions(options *Options, inout *cli.ProcInout) int {
	if options.CommonOptions.Help {
		return exitRefines
	}
	if options.CommonOptions.Version {
		fmt.Fprintln(inout.Stdout, version.Version)
		return exitRefines
	}

	abstract, err := readDiagram(options.AbstractName, options.AbstractReader)
	if err != nil {
		fmt.Fprintln(inout.Stderr, err)
		return exitError
	}
	detail, err := readDiagram(options.DetailName, options.DetailReader)
	if err != nil {
		fmt.Fprintln(inout.Stderr, err)
		return exitError
	}

	if err := pumlseqast.ValidateFolding(options.Phi, abstract, detail); err != nil {
		fmt.Fprintln(inout.Stderr, err)
		return exitError
	}

	absDetail, err := pumlseqast.FoldAll(detail, options.Phi)
	if err != nil {
		fmt.Fprintln(inout.Stderr, err)
		return exitError
	}

	result := csp.RefinesTraces(csp.Compile(absDetail), csp.Compile(abstract))
	if result.Refines {
		return exitRefines
	}
	fmt.Fprint(inout.Stderr, formatViolation(result))
	return exitNotRefines
}

func readDiagram(name string, r io.ReadCloser) (*pumlseqast.Diagram, error) {
	defer r.Close()
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	src, err := pumlpng.Source(b)
	if err != nil {
		return nil, err
	}
	d, err := pumlseqast.Parse(string(src))
	if err != nil {
		return nil, err
	}
	if err := pumlseqast.ValidateNoRef(name, d); err != nil {
		return nil, err
	}
	return d, nil
}

func formatViolation(r csp.Result) string {
	var b strings.Builder
	b.WriteString("Refinement does not hold.\n")
	b.WriteString("After the following trace, detail can perform an event that abstract does not allow:\n")
	b.WriteString("  trace: ")
	if len(r.Witness) == 0 {
		b.WriteString("<empty>")
	} else {
		parts := make([]string, len(r.Witness))
		for i, k := range r.Witness {
			parts[i] = k.String()
		}
		b.WriteString(strings.Join(parts, ", "))
	}
	b.WriteByte('\n')
	b.WriteString("  forbidden event: " + r.Offending.String() + "\n")
	return b.String()
}
