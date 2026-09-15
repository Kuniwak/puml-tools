package pumlseqtrace

import (
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/Kuniwak/puml-tools/outfiles"
	"github.com/Kuniwak/puml-tools/pumlseqast"
)

func Files(traces [][]pumlseqast.MessageKey, basename string) ([]outfiles.File, error) {
	if err := outfiles.ValidateBasename(basename); err != nil {
		return nil, err
	}

	files := make([]outfiles.File, 0, len(traces))
	for i, tr := range traces {
		files = append(files, outfiles.File{
			Name:    FileName(basename, i+1, len(traces)),
			Content: Render(tr),
		})
	}
	return files, nil
}

func FileName(basename string, idx, total int) string {
	return fmt.Sprintf("%s_trace_%0*d.tsv", basename, len(fmt.Sprint(total)), idx)
}

func Render(trace []pumlseqast.MessageKey) string {
	var b strings.Builder
	w := csv.NewWriter(&b)
	w.Comma = '\t'
	for _, k := range trace {

		_ = w.Write([]string{k.Label})
	}
	w.Flush()
	return b.String()
}
