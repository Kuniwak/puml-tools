package pumlseqexpand

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func readerOf(files map[string]string) func(string) ([]byte, error) {
	return func(path string) ([]byte, error) {
		src, ok := files[path]
		if !ok {
			return nil, fmt.Errorf("no such file: %s", path)
		}
		return []byte(src), nil
	}
}

func TestExpand(t *testing.T) {
	tests := map[string]struct {
		Src      string
		Base     string
		Files    map[string]string
		Expected string
	}{
		"returns unchanged when there is no ref": {
			Src:      "@startuml\nA -> B: x\n@enduml\n",
			Expected: "@startuml\nA -> B: x\n@enduml\n",
		},
		"replaces a single-line ref with the target's body": {
			Src: "@startuml\nparticipant A\nparticipant B\nA -> B: req\nref over A, B : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nparticipant A\nparticipant B\nA -> B: inner\nB -> A: ack\n@enduml\n",
			},
			Expected: "@startuml\nparticipant A\nparticipant B\nA -> B: req\nA -> B: inner\nB -> A: ack\n@enduml\n",
		},
		"expands a multi-line ref too": {
			Src: "@startuml\nparticipant A\nparticipant B\nref over A, B\n  ./sub.puml\nend ref\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nA -> B: inner\n@enduml\n",
			},
			Expected: "@startuml\nparticipant A\nparticipant B\nA -> B: inner\n@enduml\n",
		},
		"applies the ref's indentation to the expanded body": {
			Src: "@startuml\nparticipant A\nparticipant B\ngroup G\n  ref over A, B : ./sub.puml\nend\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nA -> B: inner\nB -> A: ack\n@enduml\n",
			},
			Expected: "@startuml\nparticipant A\nparticipant B\ngroup G\n  A -> B: inner\n  B -> A: ack\nend\n@enduml\n",
		},
		"indents the target's block structure while preserving relative nesting": {
			Src: "@startuml\nparticipant A\nparticipant B\ngroup G\n  ref over A, B : ./sub.puml\nend\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nalt c\n  A -> B: x\nelse\n  B -> A: y\nend\n@enduml\n",
			},
			Expected: "@startuml\nparticipant A\nparticipant B\ngroup G\n  alt c\n    A -> B: x\n  else\n    B -> A: y\n  end\nend\n@enduml\n",
		},
		"expands refs recursively": {
			Src: "@startuml\nparticipant A\nparticipant B\nref over A, B : ./mid.puml\n@enduml\n",
			Files: map[string]string{
				"mid.puml":  "@startuml\nA -> B: mid\nref over A, B : ./leaf.puml\n@enduml\n",
				"leaf.puml": "@startuml\nB -> A: leaf\n@enduml\n",
			},
			Expected: "@startuml\nparticipant A\nparticipant B\nA -> B: mid\nB -> A: leaf\n@enduml\n",
		},
		"resolves the target's relative paths from the target's own directory": {
			Src: "@startuml\nparticipant A\nparticipant B\nref over A, B : ./sub/mid.puml\n@enduml\n",
			Files: map[string]string{
				"sub/mid.puml":  "@startuml\nref over A, B : ./leaf.puml\n@enduml\n",
				"sub/leaf.puml": "@startuml\nA -> B: leaf\n@enduml\n",
			},
			Expected: "@startuml\nparticipant A\nparticipant B\nA -> B: leaf\n@enduml\n",
		},
		"base acts as the referencing file's directory": {
			Src:  "@startuml\nparticipant A\nparticipant B\nref over A, B : ./sub.puml\n@enduml\n",
			Base: "diagrams",
			Files: map[string]string{
				"diagrams/sub.puml": "@startuml\nA -> B: inner\n@enduml\n",
			},
			Expected: "@startuml\nparticipant A\nparticipant B\nA -> B: inner\n@enduml\n",
		},
		"appends participants declared only by the target to the end of the declaration block": {
			Src: "@startuml\nparticipant A\nparticipant B\nref over A, B, C : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nparticipant A\nparticipant \"Cache\" as C\nA -> C: get\n@enduml\n",
			},
			Expected: "@startuml\nparticipant A\nparticipant B\nparticipant \"Cache\" as C\nA -> C: get\n@enduml\n",
		},
		"does not place a hoisted declaration inside a box": {
			Src: "@startuml\nbox \"Front\"\nparticipant A\nend box\nA -> A: x\nref over A, C : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nparticipant C\nA -> C: get\n@enduml\n",
			},
			Expected: "@startuml\nparticipant C\nbox \"Front\"\nparticipant A\nend box\nA -> A: x\nA -> C: get\n@enduml\n",
		},
		"does not hoist into a box even with an extra end box": {
			Src: "@startuml\nend box\nparticipant A\nbox \"F\"\nparticipant D\nend box\nref over A, C : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nparticipant C\nA -> C: get\n@enduml\n",
			},
			Expected: "@startuml\nend box\nparticipant A\nparticipant C\nbox \"F\"\nparticipant D\nend box\nA -> C: get\n@enduml\n",
		},
		"hoists to a declaration outside any box": {
			Src: "@startuml\nbox \"Front\"\nparticipant A\nend box\nparticipant B\nref over A, B, C : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nparticipant C\nB -> C: get\n@enduml\n",
			},
			Expected: "@startuml\nbox \"Front\"\nparticipant A\nend box\nparticipant B\nparticipant C\nB -> C: get\n@enduml\n",
		},
		"hoists right after @startuml when the parent has no declarations": {
			Src: "@startuml\nref over A, B : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nparticipant A\nparticipant B\nA -> B: x\n@enduml\n",
			},
			Expected: "@startuml\nparticipant A\nparticipant B\nA -> B: x\n@enduml\n",
		},
		"places the expansion at the ref's position (even before later declarations)": {
			Src: "@startuml\nparticipant A\nref over A, B : ./sub.puml\nparticipant B\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nA -> B: inner\n@enduml\n",
			},
			Expected: "@startuml\nparticipant A\nA -> B: inner\nparticipant B\n@enduml\n",
		},
		"hoists a declaration only once even when the same target is referenced twice": {
			Src: "@startuml\nparticipant A\nref over A, C : ./sub.puml\nref over A, C : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nparticipant C\nA -> C: get\n@enduml\n",
			},
			Expected: "@startuml\nparticipant A\nparticipant C\nA -> C: get\nA -> C: get\n@enduml\n",
		},
		"inlines the target's comments, blank lines, and dividers unchanged": {
			Src: "@startuml\nparticipant A\nref over A : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\n' note to self\n\n== Phase 2 ==\nA -> A: s\n@enduml\n",
			},
			Expected: "@startuml\nparticipant A\n' note to self\n\n== Phase 2 ==\nA -> A: s\n@enduml\n",
		},
		"hoists a declaration even when the target declares it inside a block": {
			Src: "@startuml\nparticipant A\nref over A, C : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nopt maybe\n  participant C\n  A -> C: get\nend\n@enduml\n",
			},
			Expected: "@startuml\nparticipant A\nparticipant C\nopt maybe\n  A -> C: get\nend\n@enduml\n",
		},
		"drops the target's skinparam before inlining": {
			Src: "@startuml\nparticipant A\nref over A : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nskinparam monochrome true\nA -> A: s\n@enduml\n",
			},
			Expected: "@startuml\nparticipant A\nA -> A: s\n@enduml\n",
		},
		"drops the target's title, keeping the parent's title": {
			Src: "@startuml\ntitle Parent\nparticipant A\nref over A : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\ntitle Child\nA -> A: s\n@enduml\n",
			},
			Expected: "@startuml\ntitle Parent\nparticipant A\nA -> A: s\n@enduml\n",
		},
		"drops the target's autonumber, leaving the parent's numbering unchanged": {
			Src: "@startuml\nparticipant A\nref over A : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nautonumber\nA -> A: s\n@enduml\n",
			},
			Expected: "@startuml\nparticipant A\nA -> A: s\n@enduml\n",
		},
		"drops a diagram-wide directive even inside the target's block": {
			Src: "@startuml\nparticipant A\nref over A : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nopt maybe\n  skinparam monochrome true\n  A -> A: s\nend\n@enduml\n",
			},
			Expected: "@startuml\nparticipant A\nopt maybe\n  A -> A: s\nend\n@enduml\n",
		},
		"keeps a diagram-wide directive written inside a comment": {
			Src: "@startuml\nparticipant A\nref over A : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\n' skinparam monochrome true\nA -> A: s\n@enduml\n",
			},
			Expected: "@startuml\nparticipant A\n' skinparam monochrome true\nA -> A: s\n@enduml\n",
		},
		"keeps the parent diagram's own diagram-wide directive": {
			Src: "@startuml\nskinparam monochrome true\nparticipant A\nref over A : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nA -> A: s\n@enduml\n",
			},
			Expected: "@startuml\nskinparam monochrome true\nparticipant A\nA -> A: s\n@enduml\n",
		},
		"places the target's box declaration into the parent's box of the same name": {
			Src: "@startuml\nbox \"Ops\"\nparticipant A\nend box\nref over A, C : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nbox \"Ops\"\nparticipant C\nend box\nA -> C: get\n@enduml\n",
			},
			Expected: "@startuml\nbox \"Ops\"\nparticipant A\nparticipant C\nend box\nA -> C: get\n@enduml\n",
		},
		"appends the whole box to the end of the declaration block when the parent has no matching box": {
			Src: "@startuml\nparticipant A\nref over A, C : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nbox \"Back\" #888888\nparticipant C\nend box\nA -> C: get\n@enduml\n",
			},
			Expected: "@startuml\nparticipant A\nbox \"Back\" #888888\nparticipant C\nend box\nA -> C: get\n@enduml\n",
		},
		"keeps a declaration outside the target's box out of any box, preserving the target's order": {
			Src: "@startuml\nparticipant A\nref over A, C, D : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nbox \"Back\"\nparticipant C\nend box\nparticipant D\nA -> C: get\nA -> D: get\n@enduml\n",
			},
			Expected: "@startuml\nparticipant A\nbox \"Back\"\nparticipant C\nend box\nparticipant D\nA -> C: get\nA -> D: get\n@enduml\n",
		},
		"hoists nothing when the parent and target already share the same box": {
			Src: "@startuml\nbox \"Back\"\nparticipant C\nend box\nparticipant A\nref over A, C : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nbox \"Back\"\nparticipant C\nend box\nA -> C: get\n@enduml\n",
			},
			Expected: "@startuml\nbox \"Back\"\nparticipant C\nend box\nparticipant A\nA -> C: get\n@enduml\n",
		},
		"appends a box only once even when the same target is referenced twice": {
			Src: "@startuml\nparticipant A\nref over A, C : ./sub.puml\nref over A, C : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nbox \"Back\"\nparticipant C\nend box\nA -> C: get\n@enduml\n",
			},
			Expected: "@startuml\nparticipant A\nbox \"Back\"\nparticipant C\nend box\nA -> C: get\nA -> C: get\n@enduml\n",
		},
		"places into the box whose label matches, even with multiple parent boxes": {
			Src: "@startuml\nbox \"Front\"\nparticipant A\nend box\nbox \"Back\"\nparticipant B\nend box\nref over A, B, C : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nbox \"Back\"\nparticipant C\nend box\nA -> C: get\n@enduml\n",
			},
			Expected: "@startuml\nbox \"Front\"\nparticipant A\nend box\nbox \"Back\"\nparticipant B\nparticipant C\nend box\nA -> C: get\n@enduml\n",
		},
		"does not conflict when the parent declares inside a block within a box": {
			Src: "@startuml\nbox \"B\"\nopt x\nparticipant A\nend\nend box\nref over A : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nA -> A: s\n@enduml\n",
			},
			Expected: "@startuml\nbox \"B\"\nopt x\nparticipant A\nend\nend box\nA -> A: s\n@enduml\n",
		},
		"keeps the contents of a block comment": {
			Src: "@startuml\nparticipant A\nref over A : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\n/'\nskinparam monochrome true\n'/\nA -> A: s\n@enduml\n",
			},
			Expected: "@startuml\nparticipant A\n/'\nskinparam monochrome true\n'/\nA -> A: s\n@enduml\n",
		},
		"drops a multi-line diagram-wide directive along with its body": {
			Src: "@startuml\nparticipant A\nref over A : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nlegend\n  hide these rows\nend legend\nA -> A: s\n@enduml\n",
			},
			Expected: "@startuml\nparticipant A\nA -> A: s\n@enduml\n",
		},
		"drops a multi-line title along with its body": {
			Src: "@startuml\nparticipant A\nref over A : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\ntitle\n  Child\nend title\nA -> A: s\n@enduml\n",
			},
			Expected: "@startuml\nparticipant A\nA -> A: s\n@enduml\n",
		},
		"does not consume the next line for a single-line title": {
			Src: "@startuml\nparticipant A\nref over A : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\ntitle Child\n' keep me\nA -> A: s\n@enduml\n",
			},
			Expected: "@startuml\nparticipant A\n' keep me\nA -> A: s\n@enduml\n",
		},
		"carries over the target's line terminators for a hoisted box": {
			Src: "@startuml\nparticipant A\nref over A, C : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\r\nbox \"Back\"\r\nparticipant C\r\nend box\r\nA -> C: get\r\n@enduml\r\n",
			},
			Expected: "@startuml\nparticipant A\nbox \"Back\"\r\nparticipant C\r\nend box\r\nA -> C: get\r\n@enduml\n",
		},
		"also drops scale, hide, show, skin, and newpage": {
			Src: "@startuml\nparticipant A\nref over A : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nscale 2\nhide footbox\nshow A\nskin rose\nnewpage\nA -> A: s\n@enduml\n",
			},
			Expected: "@startuml\nparticipant A\nA -> A: s\n@enduml\n",
		},
		"hoists a box that specifies only a color": {
			Src: "@startuml\nparticipant A\nref over A, C : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nbox #AliceBlue\nparticipant C\nend box\nA -> C: get\n@enduml\n",
			},
			Expected: "@startuml\nparticipant A\nbox #AliceBlue\nparticipant C\nend box\nA -> C: get\n@enduml\n",
		},
		"treats boxes with different colors as distinct (hoisting after @startuml when there is no declaration outside any box)": {
			Src: "@startuml\nbox #AliceBlue\nparticipant A\nend box\nref over A, C : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nbox #Pink\nparticipant C\nend box\nA -> C: get\n@enduml\n",
			},
			Expected: "@startuml\nbox #Pink\nparticipant C\nend box\nbox #AliceBlue\nparticipant A\nend box\nA -> C: get\n@enduml\n",
		},
		"places into the same box when the label matches even if the color differs": {
			Src: "@startuml\nbox \"Ops\" #AliceBlue\nparticipant A\nend box\nref over A, C : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nbox \"Ops\" #Pink\nparticipant C\nend box\nA -> C: get\n@enduml\n",
			},
			Expected: "@startuml\nbox \"Ops\" #AliceBlue\nparticipant A\nparticipant C\nend box\nA -> C: get\n@enduml\n",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			d, _, err := Expand("seq.puml", []byte(test.Src), test.Base, readerOf(test.Files))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := d.String(); got != test.Expected {
				t.Errorf("Expand mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}

func TestExpandErrors(t *testing.T) {
	tests := map[string]struct {
		Src      string
		Files    map[string]string
		Contains string
	}{
		"errors when the label is an absolute path": {
			Src:      "@startuml\nref over A, B : /abs/sub.puml\n@enduml\n",
			Contains: "relative path",
		},
		"errors when the label is a URL": {
			Src:      "@startuml\nref over A, B : https://example.com/sub.puml\n@enduml\n",
			Contains: "relative path",
		},
		"a non-path label fails to read and reports it must be a relative path": {
			Src:      "@startuml\nref over A, B : Auth\n@enduml\n",
			Contains: "relative path",
		},
		"errors when the target cannot be read": {
			Src:      "@startuml\nref over A, B : ./missing.puml\n@enduml\n",
			Contains: "cannot read ref target: missing.puml",
		},
		"scope violation when the target uses a lifeline missing from over": {
			Src: "@startuml\nref over A, B : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nA -> B: x\nB -> Z: leak\n@enduml\n",
			},
			Contains: "lifeline \"Z\", which is not covered",
		},
		"errors when the target does not end with @enduml": {
			Src: "@startuml\nA -> B: x\nref over A, B : ./sub.puml\nA -> B: y\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nA -> B: inner",
			},
			Contains: "single diagram",
		},
		"errors when the target has no @startuml": {
			Src: "@startuml\nref over A, B : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "A -> B: inner\n",
			},
			Contains: "single diagram",
		},
		"errors when the target contains two diagrams": {
			Src: "@startuml\nref over A, B : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nA -> B: one\n@enduml\n@startuml\nA -> B: two\n@enduml\n",
			},
			Contains: "single diagram",
		},
		"errors on a circular reference": {
			Src: "@startuml\nref over A, B : ./a.puml\n@enduml\n",
			Files: map[string]string{
				"a.puml": "@startuml\nref over A, B : ./b.puml\n@enduml\n",
				"b.puml": "@startuml\nref over A, B : ./a.puml\n@enduml\n",
			},
			Contains: "cycle",
		},
		"errors when the parent and target boxes differ": {
			Src: "@startuml\nbox \"Front\"\nparticipant C\nend box\nref over C : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nbox \"Back\"\nparticipant C\nend box\nC -> C: x\n@enduml\n",
			},
			Contains: "participant \"C\"'s box differs: box \"Front\" vs box \"Back\"",
		},
		"errors when the parent is outside a box and the target is inside one": {
			Src: "@startuml\nparticipant C\nref over C : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nbox \"Back\"\nparticipant C\nend box\nC -> C: x\n@enduml\n",
			},
			Contains: "participant \"C\"'s box differs: outside any box vs box \"Back\"",
		},
		"errors when the parent is inside a box and the target is outside one": {
			Src: "@startuml\nbox \"Back\"\nparticipant C\nend box\nref over C : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nparticipant C\nC -> C: x\n@enduml\n",
			},
			Contains: "participant \"C\"'s box differs: box \"Back\" vs outside any box",
		},
		"distinguishes a box with no label from being outside any box": {
			Src: "@startuml\nparticipant C\nref over C : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nbox\nparticipant C\nend box\nC -> C: x\n@enduml\n",
			},
			Contains: "participant \"C\"'s box differs",
		},
		"errors when two targets place the same id in different boxes": {
			Src: "@startuml\nref over C : ./a.puml\nref over C : ./b.puml\n@enduml\n",
			Files: map[string]string{
				"a.puml": "@startuml\nbox \"Front\"\nparticipant C\nend box\nC -> C: a\n@enduml\n",
				"b.puml": "@startuml\nbox \"Back\"\nparticipant C\nend box\nC -> C: b\n@enduml\n",
			},
			Contains: "participant \"C\"'s box differs: box \"Front\" vs box \"Back\"",
		},
		"errors when the same id is declared with a different label": {
			Src: "@startuml\nparticipant \"Cache\" as C\nref over C : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nparticipant \"Store\" as C\nC -> C: x\n@enduml\n",
			},
			Contains: "participant \"C\" declared inconsistently",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			_, _, err := Expand("seq.puml", []byte(test.Src), "", readerOf(test.Files))
			if err == nil {
				t.Fatal("want error, got nil")
			}
			if !strings.Contains(err.Error(), test.Contains) {
				t.Errorf("error %q does not contain %q", err.Error(), test.Contains)
			}
		})
	}
}

func TestExpandPanicsOnNilRead(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Expand with a nil read should panic, but returned")
		}
	}()
	_, _, _ = Expand("seq.puml", []byte("@startuml\nref over A : ./sub.puml\n@enduml\n"), "", nil)
}

func TestExpandStripWarnings(t *testing.T) {
	tests := map[string]struct {
		Src      string
		Files    map[string]string
		Expected []StripWarning
	}{
		"no warning when nothing is dropped": {
			Src: "@startuml\nparticipant A\nref over A : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nA -> A: s\n@enduml\n",
			},
			Expected: nil,
		},
		"a dropped directive's warning points to the referencing file and the ref's line": {
			Src: "@startuml\nparticipant A\nref over A : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nskinparam monochrome true\ntitle Child\nA -> A: s\n@enduml\n",
			},
			Expected: []StripWarning{
				{Name: "seq.puml", Line: 3, Path: "sub.puml", Keyword: "skinparam"},
				{Name: "seq.puml", Line: 3, Path: "sub.puml", Keyword: "title"},
			},
		},
		"one warning regardless of how many lines repeat the same directive": {
			Src: "@startuml\nparticipant A\nref over A : ./sub.puml\n@enduml\n",
			Files: map[string]string{
				"sub.puml": "@startuml\nskinparam monochrome true\nskinparam shadowing false\nA -> A: s\n@enduml\n",
			},
			Expected: []StripWarning{
				{Name: "seq.puml", Line: 3, Path: "sub.puml", Keyword: "skinparam"},
			},
		},
		"a nested ref's warning points to the intermediate target's name and line": {
			Src: "@startuml\nparticipant A\nref over A : ./mid.puml\n@enduml\n",
			Files: map[string]string{
				"mid.puml":  "@startuml\nA -> A: mid\nref over A : ./leaf.puml\n@enduml\n",
				"leaf.puml": "@startuml\nskinparam monochrome true\nA -> A: leaf\n@enduml\n",
			},
			Expected: []StripWarning{
				{Name: "mid.puml", Line: 3, Path: "leaf.puml", Keyword: "skinparam"},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			_, warnings, err := Expand("seq.puml", []byte(test.Src), "", readerOf(test.Files))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !slices.Equal(test.Expected, warnings) {
				t.Errorf("warnings mismatch (-want +got):\n%s", cmp.Diff(test.Expected, warnings))
			}
		})
	}
}

func TestBoxTracker(t *testing.T) {
	tests := map[string]struct {
		Lines         []string
		ExpectedDepth int
		ExpectedKey   string
	}{
		"depth 0 and empty key outside any box":                  {Lines: []string{"A -> A: x\n"}, ExpectedDepth: 0, ExpectedKey: ""},
		"the label is the key inside a box":                      {Lines: []string{"box \"Front\"\n"}, ExpectedDepth: 1, ExpectedKey: "Front"},
		"the label is the key for a colored box too":             {Lines: []string{"box \"Front\" #AliceBlue\n"}, ExpectedDepth: 1, ExpectedKey: "Front"},
		"the start line is the key for a box with no label":      {Lines: []string{"box\n"}, ExpectedDepth: 1, ExpectedKey: "box"},
		"the start line is the key for a box with only a color":  {Lines: []string{"box #AliceBlue\n"}, ExpectedDepth: 1, ExpectedKey: "box #AliceBlue"},
		"end box returns to outside any box":                     {Lines: []string{"box \"Front\"\n", "end box\n"}, ExpectedDepth: 0, ExpectedKey: ""},
		"an unmatched end box does not make depth negative":      {Lines: []string{"end box\n"}, ExpectedDepth: 0, ExpectedKey: ""},
		"a box after an unmatched end box is still inside a box": {Lines: []string{"end box\n", "box \"F\"\n"}, ExpectedDepth: 1, ExpectedKey: "F"},
		"a nested box keeps the outer key":                       {Lines: []string{"box \"O\"\n", "box \"I\"\n"}, ExpectedDepth: 2, ExpectedKey: "O"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var box boxTracker
			for _, l := range test.Lines {
				box.feed(l)
			}
			if box.depth != test.ExpectedDepth || box.key != test.ExpectedKey {
				t.Errorf("boxTracker mismatch (-want +got):\n%s",
					cmp.Diff([]any{test.ExpectedDepth, test.ExpectedKey}, []any{box.depth, box.key}))
			}
		})
	}
}

func TestStripComposable(t *testing.T) {
	tests := map[string]struct {
		Input            string
		ExpectedText     string
		ExpectedKeywords []string
	}{
		"returns unchanged when there is nothing to drop": {
			Input:        "' comment\n\n== Phase ==\n",
			ExpectedText: "' comment\n\n== Phase ==\n",
		},
		"drops a diagram-wide directive": {
			Input:            "skinparam monochrome true\n' keep\n",
			ExpectedText:     "' keep\n",
			ExpectedKeywords: []string{"skinparam"},
		},
		"one keyword regardless of how many lines repeat the directive": {
			Input:            "skinparam a\nskinparam b\n",
			ExpectedText:     "",
			ExpectedKeywords: []string{"skinparam"},
		},
		"drops box open/close without warning": {
			Input:        "box \"Back\"\nend box\n' keep\n",
			ExpectedText: "' keep\n",
		},
		"passes a block comment through untouched": {
			Input:        "/'\nskinparam monochrome true\n'/\n",
			ExpectedText: "/'\nskinparam monochrome true\n'/\n",
		},
		"drops a multi-line directive along with its body": {
			Input:            "legend\n  skinparam is not a directive here\nend legend\n' keep\n",
			ExpectedText:     "' keep\n",
			ExpectedKeywords: []string{"legend"},
		},
		"does not consume the next line for a single-line directive": {
			Input:            "title Child\n' keep\n",
			ExpectedText:     "' keep\n",
			ExpectedKeywords: []string{"title"},
		},
		"keeps a directive written inside a comment": {
			Input:        "' skinparam monochrome true\n",
			ExpectedText: "' skinparam monochrome true\n",
		},
		"can drop even without a trailing line terminator": {
			Input:            "' keep\nskinparam monochrome true",
			ExpectedText:     "' keep\n",
			ExpectedKeywords: []string{"skinparam"},
		},
		"can drop with CRLF too": {
			Input:            "skinparam monochrome true\r\n' keep\r\n",
			ExpectedText:     "' keep\r\n",
			ExpectedKeywords: []string{"skinparam"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			text, keywords := stripComposable(test.Input)
			if text != test.ExpectedText {
				t.Errorf("text mismatch (-want +got):\n%s", cmp.Diff(test.ExpectedText, text))
			}
			if !slices.Equal(test.ExpectedKeywords, keywords) {
				t.Errorf("keywords mismatch (-want +got):\n%s", cmp.Diff(test.ExpectedKeywords, keywords))
			}
		})
	}
}
