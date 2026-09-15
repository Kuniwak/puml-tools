package cmd

import (
	"fmt"
	"io"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/pumlseqast"
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
	folded, err := pumlseqast.Fold(d, options.IDs, options.NewID, options.NewDesc)
	if err != nil {
		return err
	}
	folded = pumlseqast.Prune(folded)
	if options.Name != "" {
		folded = pumlseqast.SetName(folded, options.Name)
	}
	_, err = io.WriteString(inout.Stdout, folded.String())
	return err
}
