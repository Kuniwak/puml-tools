package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/pumlseqast"
	"github.com/Kuniwak/puml-tools/pumlseqexpand"
	"github.com/Kuniwak/puml-tools/tools/pumlparse/pumlpng"
	"github.com/Kuniwak/puml-tools/version"
)

const (
	exitOK    = 0
	exitError = 2
)

func MainCommandByArgs(args []string, inout *cli.ProcInout) int {
	options, err := ParseOptions(args, inout)
	if err != nil {
		fmt.Fprintln(inout.Stderr, err)
		return exitError
	}

	if err := MainCommandByOptions(options, inout); err != nil {
		fmt.Fprintln(inout.Stderr, err)
		return exitError
	}

	return exitOK
}

func MainCommandByOptions(options *Options, inout *cli.ProcInout) error {
	if options.CommonOptions.Help {
		return nil
	}

	if options.CommonOptions.Version {
		fmt.Fprintln(inout.Stdout, version.Version)
		return nil
	}

	defer options.Input.Reader.Close()
	b, err := io.ReadAll(options.Input.Reader)
	if err != nil {
		return err
	}
	src, err := pumlpng.Source(b)
	if err != nil {
		return err
	}

	d, warnings, err := pumlseqexpand.Expand(options.Input.Name, src, options.Base, readSource)
	if err != nil {
		return err
	}
	for _, w := range warnings {
		fmt.Fprintf(inout.Stderr, "%s:%d: warning: %s\n", w.Name, w.Line, w.Reason())
	}
	if options.Name != "" {
		d = pumlseqast.SetName(d, options.Name)
	}
	_, err = io.WriteString(inout.Stdout, d.String())
	return err
}

func readSource(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return pumlpng.Source(b)
}
