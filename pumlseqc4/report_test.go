package pumlseqc4

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestReportEq(t *testing.T) {
	tests := map[string]struct {
		Diff     *Diff
		Expected string
	}{
		"empty": {Diff: &Diff{}, Expected: ""},
		"all sections": {
			Diff: &Diff{
				SeqOnlyNodes: []Node{{ID: "c", Label: "C"}},
				C4OnlyNodes:  []Node{{ID: "x", Label: "X"}},
				SeqOnlyEdges: []Edge{{From: "a", To: "b"}},
				C4OnlyEdges:  []Edge{{From: "a", To: "x"}},
			},
			Expected: "participants only in SEQ (missing in C4): 1\n" +
				"  - c (\"C\")\n" +
				"elements only in C4 (missing in SEQ): 1\n" +
				"  - x (\"X\")\n" +
				"messages only in SEQ (missing in C4): 1\n" +
				"  - a -> b\n" +
				"relationships only in C4 (missing in SEQ): 1\n" +
				"  - a -> x\n",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := test.Diff.ReportEq(); got != test.Expected {
				t.Errorf("ReportEq mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}

func TestReportSubset(t *testing.T) {
	tests := map[string]struct {
		Diff     *Diff
		Expected string
	}{
		"empty": {Diff: &Diff{}, Expected: ""},
		"ignores seq-only, lists c4-only": {
			Diff: &Diff{
				SeqOnlyNodes: []Node{{ID: "c", Label: "C"}},
				C4OnlyNodes:  []Node{{ID: "x", Label: "X"}},
				C4OnlyEdges:  []Edge{{From: "a", To: "x"}},
			},
			Expected: "elements only in C4 (not in SEQ): 1\n" +
				"  - x (\"X\")\n" +
				"relationships only in C4 (not in SEQ): 1\n" +
				"  - a -> x\n",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := test.Diff.ReportSubset(); got != test.Expected {
				t.Errorf("ReportSubset mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}

func TestReportSuperset(t *testing.T) {
	tests := map[string]struct {
		Diff     *Diff
		Expected string
	}{
		"empty": {Diff: &Diff{}, Expected: ""},
		"ignores c4-only, lists seq-only": {
			Diff: &Diff{
				SeqOnlyNodes: []Node{{ID: "c", Label: "C"}},
				SeqOnlyEdges: []Edge{{From: "a", To: "b"}},
				C4OnlyNodes:  []Node{{ID: "x", Label: "X"}},
				C4OnlyEdges:  []Edge{{From: "a", To: "x"}},
			},
			Expected: "participants only in SEQ (not in C4): 1\n" +
				"  - c (\"C\")\n" +
				"messages only in SEQ (not in C4): 1\n" +
				"  - a -> b\n",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := test.Diff.ReportSuperset(); got != test.Expected {
				t.Errorf("ReportSuperset mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}
