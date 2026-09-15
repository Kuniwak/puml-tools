package cmd

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/pumlseqast"
	"github.com/Kuniwak/puml-tools/sets"
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

	set := sets.New(options.Common...)
	w := csv.NewWriter(inout.Stdout)
	w.Comma = '\t'
	seen := map[pumlseqast.MessageKey]bool{}
	for _, m := range pumlseqast.Messages(d) {
		k := m.Key()
		if !set.Contains(k.Sender) || !set.Contains(k.Receiver) {
			continue
		}
		if seen[k] {
			continue
		}
		seen[k] = true
		if err := w.Write([]string{k.Sender, k.Receiver, k.Label}); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}
