package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/todo"
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

	w := csv.NewWriter(inout.Stdout)
	w.Comma = '\t'
	for _, t := range todo.Extract(src, options.Pattern) {
		if err := w.Write([]string{options.Name, strconv.Itoa(t.Line), t.Text}); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}
