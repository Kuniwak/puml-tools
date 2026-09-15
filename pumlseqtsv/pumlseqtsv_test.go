package pumlseqtsv_test

import (
	"maps"
	"strings"
	"testing"

	"github.com/Kuniwak/puml-tools/pumlseqast"
	"github.com/Kuniwak/puml-tools/pumlseqtsv"
)

func TestParseMessageKeys(t *testing.T) {
	tests := map[string]struct {
		Input       string
		Expected    map[pumlseqast.MessageKey]bool
		ExpectedErr bool
	}{
		"empty input yields an empty set": {
			Input:    "",
			Expected: map[pumlseqast.MessageKey]bool{},
		},
		"three columns are sender, receiver, and label": {
			Input: "A\tB\tlabel\n",
			Expected: map[pumlseqast.MessageKey]bool{
				{Sender: "A", Receiver: "B", Label: "label"}: true,
			},
		},
		"a missing column becomes an empty string": {
			Input: "A\tB\n",
			Expected: map[pumlseqast.MessageKey]bool{
				{Sender: "A", Receiver: "B", Label: ""}: true,
			},
		},
		"a fully-blank record is ignored": {
			Input: "A\tB\tlabel\n\t\t\n",
			Expected: map[pumlseqast.MessageKey]bool{
				{Sender: "A", Receiver: "B", Label: "label"}: true,
			},
		},
		"leading and trailing spaces are preserved": {
			Input: " A \tB\tlabel\n",
			Expected: map[pumlseqast.MessageKey]bool{
				{Sender: " A ", Receiver: "B", Label: "label"}: true,
			},
		},
		"duplicate records are merged into one": {
			Input: "A\tB\tlabel\nA\tB\tlabel\n",
			Expected: map[pumlseqast.MessageKey]bool{
				{Sender: "A", Receiver: "B", Label: "label"}: true,
			},
		},
		"malformed quoting in the TSV is an error": {
			Input:       "A\t\"B\tlabel\n",
			ExpectedErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := pumlseqtsv.ParseMessageKeys(strings.NewReader(test.Input))
			if test.ExpectedErr {
				if err == nil {
					t.Fatalf("ParseMessageKeys(%q) should fail, but got no error", test.Input)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseMessageKeys(%q) should succeed, but got error: %s", test.Input, err)
			}
			if !maps.Equal(test.Expected, got) {
				t.Errorf("ParseMessageKeys(%q) mismatch: want %v, got %v", test.Input, test.Expected, got)
			}
		})
	}
}

func TestParseFolding(t *testing.T) {
	tests := map[string]struct {
		Input       string
		Expected    pumlseqast.Folding
		ExpectedErr bool
	}{
		"empty input yields an empty mapping": {
			Input:    "",
			Expected: pumlseqast.Folding{},
		},
		"two columns are the detail id and the abstract id": {
			Input:    "B\tBE\nC\tBE\n",
			Expected: pumlseqast.Folding{"B": "BE", "C": "BE"},
		},
		"a fully-blank record is ignored": {
			Input:    "B\tBE\n\t\n",
			Expected: pumlseqast.Folding{"B": "BE"},
		},
		"leading and trailing spaces are preserved": {
			Input:    " B \tBE\n",
			Expected: pumlseqast.Folding{" B ": "BE"},
		},
		"columns after the second are ignored": {
			Input:    "B\tBE\textra\n",
			Expected: pumlseqast.Folding{"B": "BE"},
		},
		"a later row overwrites the same detail id": {
			Input:    "B\tBE\nB\tFE\n",
			Expected: pumlseqast.Folding{"B": "FE"},
		},
		"a record missing the abstract id is an error": {
			Input:       "B\n",
			ExpectedErr: true,
		},
		"a record missing the detail id is an error": {
			Input:       "\tBE\n",
			ExpectedErr: true,
		},
		"malformed quoting in the TSV is an error": {
			Input:       "B\t\"BE\n",
			ExpectedErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := pumlseqtsv.ParseFolding(strings.NewReader(test.Input))
			if test.ExpectedErr {
				if err == nil {
					t.Fatalf("ParseFolding(%q) should fail, but got no error", test.Input)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseFolding(%q) should succeed, but got error: %s", test.Input, err)
			}
			if !maps.Equal(test.Expected, got) {
				t.Errorf("ParseFolding(%q) mismatch: want %v, got %v", test.Input, test.Expected, got)
			}
		})
	}
}
