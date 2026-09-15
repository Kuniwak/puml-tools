package pumlseqtrace

import (
	"testing"

	"github.com/Kuniwak/puml-tools/outfiles"
	"github.com/Kuniwak/puml-tools/pumlseqast"
	"github.com/google/go-cmp/cmp"
)

func key(label string) pumlseqast.MessageKey {
	return pumlseqast.MessageKey{Sender: "A", Receiver: "B", Label: label}
}

func TestFiles(t *testing.T) {
	tests := map[string]struct {
		Traces      [][]pumlseqast.MessageKey
		Basename    string
		Expected    []outfiles.File
		ExpectedErr error
	}{
		"one file per trace, numbered from 1": {
			Traces:   [][]pumlseqast.MessageKey{{key("x"), key("y")}, {key("z")}},
			Basename: "seq",
			Expected: []outfiles.File{
				{Name: "seq_trace_1.tsv", Content: "x\ny\n"},
				{Name: "seq_trace_2.tsv", Content: "z\n"},
			},
		},
		"the index is zero padded to the digits of the total": {
			Traces:   [][]pumlseqast.MessageKey{{key("x")}, {}, {}, {}, {}, {}, {}, {}, {}, {}},
			Basename: "seq",
			Expected: []outfiles.File{
				{Name: "seq_trace_01.tsv", Content: "x\n"},
				{Name: "seq_trace_02.tsv"},
				{Name: "seq_trace_03.tsv"},
				{Name: "seq_trace_04.tsv"},
				{Name: "seq_trace_05.tsv"},
				{Name: "seq_trace_06.tsv"},
				{Name: "seq_trace_07.tsv"},
				{Name: "seq_trace_08.tsv"},
				{Name: "seq_trace_09.tsv"},
				{Name: "seq_trace_10.tsv"},
			},
		},
		"a label containing a tab stays one field": {
			Traces:   [][]pumlseqast.MessageKey{{key("a\tb")}},
			Basename: "seq",
			Expected: []outfiles.File{{Name: "seq_trace_1.tsv", Content: "\"a\tb\"\n"}},
		},
		"a label containing a newline stays one record": {
			Traces:   [][]pumlseqast.MessageKey{{key("a\nb"), key("c")}},
			Basename: "seq",
			Expected: []outfiles.File{{Name: "seq_trace_1.tsv", Content: "\"a\nb\"\nc\n"}},
		},
		"a label containing a double quote is escaped": {
			Traces:   [][]pumlseqast.MessageKey{{key(`a"b`)}},
			Basename: "seq",
			Expected: []outfiles.File{{Name: "seq_trace_1.tsv", Content: "\"a\"\"b\"\n"}},
		},
		"an empty trace makes an empty file": {
			Traces:   [][]pumlseqast.MessageKey{{}},
			Basename: "seq",
			Expected: []outfiles.File{{Name: "seq_trace_1.tsv", Content: ""}},
		},
		"a basename escaping the out dir is rejected": {
			Traces:      [][]pumlseqast.MessageKey{{key("x")}},
			Basename:    "../escaped",
			ExpectedErr: outfiles.ErrBasenameHasPathSeparator,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := Files(test.Traces, test.Basename)
			if test.ExpectedErr != nil {
				if err == nil || !cmp.Equal(err.Error(), test.ExpectedErr.Error()) {
					t.Fatalf("Files returned error %v, want %v", err, test.ExpectedErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Files returned error: %s", err)
			}
			if diff := cmp.Diff(test.Expected, got); diff != "" {
				t.Errorf("Files mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
