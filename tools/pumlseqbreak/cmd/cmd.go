package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/pumlseqbreak"
	"github.com/Kuniwak/puml-tools/tools/pumlparse/pumlpng"
	"github.com/Kuniwak/puml-tools/version"
)

func MainCommandByArgs(args []string, inout *cli.ProcInout) int {
	options, err := ParseOptions(args, inout)
	if err != nil {
		fmt.Fprintln(inout.Stderr, err)
		return 1
	}

	found, err := MainCommandByOptions(options, inout)
	if err != nil {
		fmt.Fprintln(inout.Stderr, err)
		return 1
	}

	if found > 0 {
		return 1
	}
	return 0
}

func MainCommandByOptions(options *Options, inout *cli.ProcInout) (int, error) {
	if options.CommonOptions.Help {
		return 0, nil
	}

	if options.CommonOptions.Version {
		fmt.Fprintln(inout.Stdout, version.Version)
		return 0, nil
	}

	defer options.Reader.Close()
	b, err := io.ReadAll(options.Reader)
	if err != nil {
		return 0, err
	}

	src, err := pumlpng.Source(b)
	if err != nil {
		return 0, err
	}

	breaks := pumlseqbreak.Find(src)
	w := csv.NewWriter(inout.Stdout)
	w.Comma = '\t'
	for _, br := range breaks {
		if err := w.Write([]string{options.Name, strconv.Itoa(br.Line), br.Reason()}); err != nil {
			return 0, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return 0, err
	}
	return len(breaks), nil
}
