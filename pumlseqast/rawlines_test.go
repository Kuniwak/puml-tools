package pumlseqast

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNodeLines(t *testing.T) {
	tests := map[string]struct {
		Nodes    []Node
		Expected []NodeLine
	}{
		"expands a raw node into its physical lines": {
			Nodes:    []Node{&Raw{Text: "a\nb\n"}},
			Expected: []NodeLine{{Node: 0, Line: 0, Text: "a"}, {Node: 0, Line: 1, Text: "b"}},
		},
		"typed nodes are one item with Line -1": {
			Nodes:    []Node{&Participant{Text: "participant A", Term: "\n"}},
			Expected: []NodeLine{{Node: 0, Line: -1}},
		},
		"mixes raw lines and typed nodes in order": {
			Nodes: []Node{
				&Raw{Text: "box \"G\"\n"},
				&Participant{Text: "participant A", Term: "\n"},
				&Raw{Text: "end box\n"},
			},
			Expected: []NodeLine{
				{Node: 0, Line: 0, Text: "box \"G\""},
				{Node: 1, Line: -1},
				{Node: 2, Line: 0, Text: "end box"},
			},
		},
		"last line without a trailing newline": {
			Nodes:    []Node{&Raw{Text: "a\nb"}},
			Expected: []NodeLine{{Node: 0, Line: 0, Text: "a"}, {Node: 0, Line: 1, Text: "b"}},
		},
		"no nodes": {
			Nodes:    nil,
			Expected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if diff := cmp.Diff(test.Expected, NodeLines(test.Nodes)); diff != "" {
				t.Errorf("NodeLines mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDropRawLines(t *testing.T) {
	tests := map[string]struct {
		Nodes    []Node
		Drop     []string
		Expected string
	}{
		"drops the matching raw line": {
			Nodes:    []Node{&Raw{Text: "a\nb\n"}},
			Drop:     []string{"a"},
			Expected: "b\n",
		},
		"drops the node that lost every line": {
			Nodes:    []Node{&Raw{Text: "a\n"}, &Participant{Text: "participant A", Term: "\n"}},
			Drop:     []string{"a"},
			Expected: "participant A\n",
		},
		"keeps typed nodes whatever the predicate says": {
			Nodes:    []Node{&Participant{Text: "participant A", Term: "\n"}},
			Drop:     []string{"", "participant A"},
			Expected: "participant A\n",
		},
		"keeps blank lines and comments around a dropped line": {
			Nodes:    []Node{&Raw{Text: "box \"G\"\n' memo\n\nend box\n"}},
			Drop:     []string{"box \"G\"", "end box"},
			Expected: "' memo\n\n",
		},
		"last line without a trailing newline": {
			Nodes:    []Node{&Raw{Text: "a\nb"}},
			Drop:     []string{"a"},
			Expected: "b",
		},
		"CRLF line terminators are preserved": {
			Nodes:    []Node{&Raw{Text: "a\r\nb\r\n"}},
			Drop:     []string{"a"},
			Expected: "b\r\n",
		},
		"empty predicate is the identity": {
			Nodes:    []Node{&Raw{Text: "a\nb\n"}},
			Drop:     nil,
			Expected: "a\nb\n",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			before := (&Diagram{Nodes: test.Nodes}).String()
			out := DropRawLines(test.Nodes, func(nl NodeLine) bool {
				for _, d := range test.Drop {
					if strings.TrimSpace(nl.Text) == d {
						return true
					}
				}
				return false
			})
			if diff := cmp.Diff(test.Expected, (&Diagram{Nodes: out}).String()); diff != "" {
				t.Errorf("DropRawLines mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(before, (&Diagram{Nodes: test.Nodes}).String()); diff != "" {
				t.Errorf("input was mutated (-want +got):\n%s", diff)
			}
		})
	}
}
