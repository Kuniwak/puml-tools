package cmd

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/pumlseqast"
	"github.com/Kuniwak/puml-tools/pumlseqsubset"
	"github.com/Kuniwak/puml-tools/tools/pumlparse/pumlpng"
	"github.com/Kuniwak/puml-tools/version"
)

func MainCommandByArgs(args []string, inout *cli.ProcInout) int {
	options, err := ParseOptions(args, inout)
	if err != nil {
		fmt.Fprintln(inout.Stderr, err)
		return 2
	}

	residuals, err := MainCommandByOptions(options, inout)
	if err != nil {
		fmt.Fprintln(inout.Stderr, err)
		return 2
	}

	if residuals > 0 {
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

	defer options.SuperReader.Close()
	defer options.SubReader.Close()

	superSeq, err := readSeq(options.SuperName, options.SuperReader)
	if err != nil {
		return 0, err
	}
	subSeq, err := readSeq(options.SubName, options.SubReader)
	if err != nil {
		return 0, err
	}

	mode, err := pumlseqsubset.ParseMode(string(options.Mode))
	if err != nil {
		return 0, err
	}
	residuals := pumlseqsubset.Residuals(mode, subSeq, superSeq)

	w := csv.NewWriter(inout.Stdout)
	w.Comma = '\t'
	for _, r := range residuals {
		if err := w.Write(pumlseqsubset.Record(mode, r)); err != nil {
			return 0, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return 0, err
	}
	return len(residuals), nil
}

func readSeq(name string, r io.Reader) (pumlseqsubset.Seq, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return pumlseqsubset.Seq{}, err
	}

	src, err := pumlpng.Source(b)
	if err != nil {
		return pumlseqsubset.Seq{}, err
	}

	d, err := pumlseqast.Parse(string(src))
	if err != nil {
		return pumlseqsubset.Seq{}, err
	}
	if err := pumlseqast.ValidateNoRef(name, d); err != nil {
		return pumlseqsubset.Seq{}, err
	}
	return pumlseqsubset.SeqOfDiagram(d), nil
}
