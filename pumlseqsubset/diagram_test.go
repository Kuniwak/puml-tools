package pumlseqsubset

import (
	"reflect"
	"testing"

	"github.com/Kuniwak/puml-tools/pumlseqast"
	"github.com/google/go-cmp/cmp"
)

func TestSeqOfDiagram(t *testing.T) {
	tests := map[string]struct {
		Src      string
		Expected Seq
	}{
		"messages only": {
			Src:      "@startuml\nA -> B\nB -> C\n@enduml\n",
			Expected: NewSeq(Pair{"A", "B"}, Pair{"B", "C"}),
		},
		"a diagram with no messages yields an empty Seq": {
			Src:      "@startuml\nparticipant A\n@enduml\n",
			Expected: NewSeq(),
		},
		"a block with no messages becomes an empty branch": {
			Src:      "@startuml\npar\nelse\nA -> B\nend\n@enduml\n",
			Expected: Seq{Items: []Item{NewPar(NewSeq(), NewSeq(Pair{"A", "B"}))}},
		},
		"participant and note are excluded": {
			Src:      "@startuml\nparticipant A\nA -> B\nnote over A: memo\n@enduml\n",
			Expected: NewSeq(Pair{"A", "B"}),
		},
		"par becomes parallel branches": {
			Src:      "@startuml\npar\nA -> B\nelse\nA -> C\nend\n@enduml\n",
			Expected: Seq{Items: []Item{NewPar(NewSeq(Pair{"A", "B"}), NewSeq(Pair{"A", "C"}))}},
		},
		"alt is flattened in written order": {
			Src:      "@startuml\nalt cond\nA -> B\nelse\nA -> C\nend\n@enduml\n",
			Expected: NewSeq(Pair{"A", "B"}, Pair{"A", "C"}),
		},
		"group is also flattened in written order": {
			Src:      "@startuml\ngroup Auth\nA -> B\nend\nB -> A\n@enduml\n",
			Expected: NewSeq(Pair{"A", "B"}, Pair{"B", "A"}),
		},
		"alt inside par is flattened": {
			Src: "@startuml\npar\nalt cond\nA -> B\nelse\nA -> C\nend\nelse\nA -> D\nend\n@enduml\n",
			Expected: Seq{Items: []Item{NewPar(
				NewSeq(Pair{"A", "B"}, Pair{"A", "C"}),
				NewSeq(Pair{"A", "D"}),
			)}},
		},
		"nested par stays nested as parallel branches": {
			Src: "@startuml\npar\nX -> Y\nelse\npar\nA -> B\nelse\nC -> D\nend\nend\n@enduml\n",
			Expected: Seq{Items: []Item{NewPar(
				NewSeq(Pair{"X", "Y"}),
				Seq{Items: []Item{NewPar(NewSeq(Pair{"A", "B"}), NewSeq(Pair{"C", "D"}))}},
			)}},
		},
		"sequential parts before and after par are preserved": {
			Src: "@startuml\nS -> A\npar\nA -> B\nelse\nA -> C\nend\nB -> S\n@enduml\n",
			Expected: Seq{Items: []Item{
				{Pair: Pair{"S", "A"}},
				NewPar(NewSeq(Pair{"A", "B"}), NewSeq(Pair{"A", "C"})),
				{Pair: Pair{"B", "S"}},
			}},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			d, err := pumlseqast.Parse(test.Src)
			if err != nil {
				t.Fatalf("Parse returned error: %v", err)
			}

			got := SeqOfDiagram(d)
			if !reflect.DeepEqual(test.Expected, got) {
				t.Errorf("SeqOfDiagram mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}

func TestSeqOfNodesPanicsOnRef(t *testing.T) {
	tests := map[string]string{
		"single-line ref":    "@startuml\nA -> B: x\nref over A, B : ./sub.puml\n@enduml\n",
		"multi-line ref":     "@startuml\nref over A, B\n  ./sub.puml\nend ref\n@enduml\n",
		"ref inside a block": "@startuml\nalt c\n  ref over A, B : ./sub.puml\nend\n@enduml\n",
	}

	for name, src := range tests {
		t.Run(name, func(t *testing.T) {
			d, err := pumlseqast.Parse(src)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer func() {
				if recover() == nil {
					t.Error("SeqOfDiagram should panic on an unexpanded ref, but returned")
				}
			}()
			SeqOfDiagram(d)
		})
	}
}
