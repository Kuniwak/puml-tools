package cmd

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/Kuniwak/puml-tools/c4ast"
	"github.com/Kuniwak/puml-tools/cli"
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

	ids := options.IDs
	if options.IDsTSVReader != nil {
		defer options.IDsTSVReader.Close()
		var err error
		ids, err = readIDsTSV(options.IDsTSVReader)
		if err != nil {
			return err
		}
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

	d, err := c4ast.Parse(string(src))
	if err != nil {
		return err
	}

	folded, err := c4ast.Fold(d, ids, options.NewID, options.NewDesc)
	if err != nil {
		return err
	}
	if options.Name != "" {
		folded.Name = options.Name
	}

	_, err = io.WriteString(inout.Stdout, folded.String())
	return err
}

func readIDsTSV(r io.Reader) ([]string, error) {
	cr := csv.NewReader(r)
	cr.Comma = '\t'
	cr.FieldsPerRecord = -1
	records, err := cr.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("ids-tsv: empty input")
	}

	idCol := -1
	for i, h := range records[0] {
		if h == "id" {
			idCol = i
			break
		}
	}
	if idCol < 0 {
		return nil, fmt.Errorf("ids-tsv: header has no %q column", "id")
	}

	var ids []string
	for _, rec := range records[1:] {
		if idCol >= len(rec) {
			continue
		}
		if id := rec[idCol]; id != "" {
			ids = append(ids, id)
		}
	}
	return ids, nil
}
