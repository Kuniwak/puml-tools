package pumlseqc4

import (
	"errors"
	"reflect"
	"testing"

	"github.com/Kuniwak/puml-tools/c4ast"
	"github.com/Kuniwak/puml-tools/pumlseqast"
	"github.com/google/go-cmp/cmp"
)

func TestCompare(t *testing.T) {
	tests := map[string]struct {
		Seq          string
		C4           string
		WantDiff     *Diff
		WantEqual    bool
		WantSubset   bool
		WantSuperset bool
	}{
		"equal": {
			Seq:          "@startuml\na -> b : hi\n@enduml\n",
			C4:           "@startuml\nSystem(a, \"a\")\nSystem(b, \"b\")\nRel(a, b, \"x\")\n@enduml\n",
			WantDiff:     &Diff{},
			WantEqual:    true,
			WantSubset:   true,
			WantSuperset: true,
		},
		"seq has extra node": {
			Seq:          "@startuml\nparticipant c\na -> b : hi\n@enduml\n",
			C4:           "@startuml\nSystem(a, \"a\")\nSystem(b, \"b\")\nRel(a, b, \"x\")\n@enduml\n",
			WantDiff:     &Diff{SeqOnlyNodes: []Node{{ID: "c", Label: "c"}}},
			WantEqual:    false,
			WantSubset:   true,
			WantSuperset: false,
		},
		"seq has extra edge": {
			Seq: "@startuml\na -> b : hi\nb -> a : re\n@enduml\n",
			C4:  "@startuml\nSystem(a, \"a\")\nSystem(b, \"b\")\nRel(a, b, \"x\")\n@enduml\n",
			WantDiff: &Diff{
				SeqOnlyEdges: []Edge{{From: "b", To: "a"}},
			},
			WantEqual:    false,
			WantSubset:   true,
			WantSuperset: false,
		},
		"c4 has extra node and edge": {
			Seq: "@startuml\na -> b : hi\n@enduml\n",
			C4:  "@startuml\nSystem(a, \"a\")\nSystem(b, \"b\")\nSystem(x, \"x\")\nRel(a, b, \"r\")\nRel(a, x, \"r\")\n@enduml\n",
			WantDiff: &Diff{
				C4OnlyNodes: []Node{{ID: "x", Label: "x"}},
				C4OnlyEdges: []Edge{{From: "a", To: "x"}},
			},
			WantEqual:    false,
			WantSubset:   false,
			WantSuperset: true,
		},
		"self edges ignored": {
			Seq:          "@startuml\na -> a : loop\na -> b : x\n@enduml\n",
			C4:           "@startuml\nSystem(a, \"a\")\nSystem(b, \"b\")\nRel(a, b, \"r\")\n@enduml\n",
			WantDiff:     &Diff{},
			WantEqual:    true,
			WantSubset:   true,
			WantSuperset: true,
		},
		"labels ignored on edges": {
			Seq:          "@startuml\na -> b : hi\n@enduml\n",
			C4:           "@startuml\nSystem(a, \"a\")\nSystem(b, \"b\")\nRel(a, b, \"totally different\")\n@enduml\n",
			WantDiff:     &Diff{},
			WantEqual:    true,
			WantSubset:   true,
			WantSuperset: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			seqD, err := pumlseqast.Parse(test.Seq)
			if err != nil {
				t.Fatal(err)
			}
			c4D, err := c4ast.Parse(test.C4)
			if err != nil {
				t.Fatal(err)
			}
			got := Compare(seqD, c4D)
			if !reflect.DeepEqual(test.WantDiff, got) {
				t.Errorf("Compare mismatch (-want +got):\n%s", cmp.Diff(test.WantDiff, got))
			}
			if got.IsEqual() != test.WantEqual {
				t.Errorf("IsEqual() = %v, want %v", got.IsEqual(), test.WantEqual)
			}
			if got.IsSubset() != test.WantSubset {
				t.Errorf("IsSubset() = %v, want %v", got.IsSubset(), test.WantSubset)
			}
			if got.IsSuperset() != test.WantSuperset {
				t.Errorf("IsSuperset() = %v, want %v", got.IsSuperset(), test.WantSuperset)
			}
		})
	}
}

func TestCompareSeqs(t *testing.T) {
	tests := map[string]struct {
		Seqs         []string
		C4           string
		WantConflict *ParticipantConflict
		WantDiff     *Diff
		WantEqual    bool
	}{
		"union of two diagrams equals c4": {
			Seqs:      []string{"@startuml\na -> b : x\n@enduml\n", "@startuml\nb -> c : y\n@enduml\n"},
			C4:        "@startuml\nSystem(a, \"a\")\nSystem(b, \"b\")\nSystem(c, \"c\")\nRel(a, b, \"x\")\nRel(b, c, \"y\")\n@enduml\n",
			WantDiff:  &Diff{},
			WantEqual: true,
		},
		"union misses a c4 element": {
			Seqs: []string{"@startuml\na -> b : x\n@enduml\n", "@startuml\nb -> c : y\n@enduml\n"},
			C4:   "@startuml\nSystem(a, \"a\")\nSystem(b, \"b\")\nSystem(c, \"c\")\nSystem(z, \"z\")\nRel(a, b, \"x\")\nRel(b, c, \"y\")\n@enduml\n",
			WantDiff: &Diff{
				C4OnlyNodes: []Node{{ID: "z", Label: "z"}},
			},
			WantEqual: false,
		},
		"single diagram behaves like Compare": {
			Seqs:      []string{"@startuml\na -> b : x\n@enduml\n"},
			C4:        "@startuml\nSystem(a, \"a\")\nSystem(b, \"b\")\nRel(a, b, \"x\")\n@enduml\n",
			WantDiff:  &Diff{},
			WantEqual: true,
		},
		"same id declared with different kinds conflicts": {
			Seqs: []string{"@startuml\nactor a\na -> b : x\n@enduml\n", "@startuml\nparticipant a\na -> c : y\n@enduml\n"},
			C4:   "@startuml\nSystem(a, \"a\")\n@enduml\n",
			WantConflict: &ParticipantConflict{
				ID:         "a",
				FirstIndex: 0,
				FirstLabel: "a",
				FirstKind:  KindPerson,
				Index:      1,
				Label:      "a",
				Kind:       KindSystem,
			},
		},
		"same id declared with different labels conflicts": {
			Seqs: []string{"@startuml\nparticipant \"A\" as a\na -> b : x\n@enduml\n", "@startuml\nparticipant \"AA\" as a\na -> c : y\n@enduml\n"},
			C4:   "@startuml\nSystem(a, \"A\")\n@enduml\n",
			WantConflict: &ParticipantConflict{
				ID:         "a",
				FirstIndex: 0,
				FirstLabel: "A",
				FirstKind:  KindSystem,
				Index:      1,
				Label:      "AA",
				Kind:       KindSystem,
			},
		},
		"implicit participant does not conflict with a declared one": {
			Seqs:      []string{"@startuml\nactor a\na -> b : x\n@enduml\n", "@startuml\na -> b : y\n@enduml\n"},
			C4:        "@startuml\nPerson(a, \"a\")\nSystem(b, \"b\")\nRel(a, b, \"x\")\n@enduml\n",
			WantDiff:  &Diff{},
			WantEqual: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			seqDs := make([]*pumlseqast.Diagram, 0, len(test.Seqs))
			for _, s := range test.Seqs {
				d, err := pumlseqast.Parse(s)
				if err != nil {
					t.Fatal(err)
				}
				seqDs = append(seqDs, d)
			}
			c4D, err := c4ast.Parse(test.C4)
			if err != nil {
				t.Fatal(err)
			}

			got, err := CompareSeqs(seqDs, c4D)
			if test.WantConflict != nil {
				var conflict *ParticipantConflict
				if !errors.As(err, &conflict) {
					t.Fatalf("error = %v, want *ParticipantConflict", err)
				}
				if !reflect.DeepEqual(test.WantConflict, conflict) {
					t.Errorf("conflict mismatch (-want +got):\n%s", cmp.Diff(test.WantConflict, conflict))
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(test.WantDiff, got) {
				t.Errorf("CompareSeqs mismatch (-want +got):\n%s", cmp.Diff(test.WantDiff, got))
			}
			if got.IsEqual() != test.WantEqual {
				t.Errorf("IsEqual() = %v, want %v", got.IsEqual(), test.WantEqual)
			}
		})
	}
}
