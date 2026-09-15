package cmd

import (
	"fmt"
	"io"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/outfiles"
	"github.com/Kuniwak/puml-tools/pumlseqast"
	"github.com/Kuniwak/puml-tools/pumlseqgroupsplit"
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

	files, err := pumlseqgroupsplit.Split(d, options.Basename)
	if err != nil {
		return err
	}

	paths, writeErr := outfiles.WriteFiles(options.OutDir, files)
	for _, path := range paths {
		fmt.Fprintln(inout.Stdout, path)
	}
	return writeErr
}
