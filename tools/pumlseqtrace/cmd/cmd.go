package cmd

import (
	"fmt"
	"io"
	"math/rand/v2"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/outfiles"
	"github.com/Kuniwak/puml-tools/pumlseqast"
	"github.com/Kuniwak/puml-tools/pumlseqtrace"
	"github.com/Kuniwak/puml-tools/tools/pumlparse/pumlpng"
	"github.com/Kuniwak/puml-tools/version"
)

const warnLimit = 100

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

	defer options.Reader.Close()
	b, err := io.ReadAll(options.Reader)
	if err != nil {
		return err
	}

	src, err := pumlpng.Source(b)
	if err != nil {
		return err
	}

	d, err := pumlseqast.Parse(string(src))
	if err != nil {
		return err
	}
	if err := pumlseqast.ValidateNoRef(options.Name, d); err != nil {
		return err
	}

	if options.Sampling == nil && pumlseqtrace.CountAtMost(d, options.Trace, warnLimit+1) > warnLimit {
		fmt.Fprintf(inout.Stderr, "warning: the trace count exceeds %d. To sample it down, interrupt with Ctrl+C and add -interleave-sampling <rate>\n", warnLimit)
	}

	traces := collect(d, options)

	files, err := pumlseqtrace.Files(traces, options.Basename)
	if err != nil {
		return err
	}

	paths, writeErr := outfiles.WriteFiles(options.OutDir, files)
	for _, path := range paths {
		fmt.Fprintln(inout.Stdout, path)
	}
	return writeErr
}

func collect(d *pumlseqast.Diagram, options *Options) [][]pumlseqast.MessageKey {
	var traces [][]pumlseqast.MessageKey
	yield := func(tr []pumlseqast.MessageKey) bool {
		traces = append(traces, tr)
		return true
	}
	if options.Sampling != nil {
		yield = pumlseqtrace.Sampled(*options.Sampling, rand.New(rand.NewPCG(options.Seed, options.Seed)), yield)
	}
	pumlseqtrace.Walk(d, options.Trace, yield)
	return traces
}
