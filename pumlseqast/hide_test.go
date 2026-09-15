package pumlseqast

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func hidePred(keys ...MessageKey) func(*Message) bool {
	set := make(map[MessageKey]bool, len(keys))
	for _, k := range keys {
		set[k] = true
	}
	return func(m *Message) bool { return set[m.Key()] }
}

func TestHide(t *testing.T) {
	tests := map[string]struct {
		Src      string
		Keys     []MessageKey
		Expected string
	}{
		"removes matching key": {
			Src:      "@startuml\nA -> A: noise\nA -> B: keep\n@enduml\n",
			Keys:     []MessageKey{{"A", "A", "noise"}},
			Expected: "@startuml\nA -> B: keep\n@enduml\n",
		},
		"non-self message removed when in set": {
			Src:      "@startuml\nA -> B: x\n@enduml\n",
			Keys:     []MessageKey{{"A", "B", "x"}},
			Expected: "@startuml\n@enduml\n",
		},
		"all three fields must match (self vs non-self distinguished)": {
			Src:      "@startuml\nA -> A: noise\nA -> B: noise\n@enduml\n",
			Keys:     []MessageKey{{"A", "A", "noise"}},
			Expected: "@startuml\nA -> B: noise\n@enduml\n",
		},
		"recurses then and else clauses": {
			Src:      "@startuml\nalt c\n  A -> A: drop\nelse\n  B -> B: drop\nend\n@enduml\n",
			Keys:     []MessageKey{{"A", "A", "drop"}, {"B", "B", "drop"}},
			Expected: "@startuml\nalt c\nelse\nend\n@enduml\n",
		},
		"ref body not affected": {
			Src:      "@startuml\nref over A\n  A -> A: drop\nend ref\n@enduml\n",
			Keys:     []MessageKey{{"A", "A", "drop"}},
			Expected: "@startuml\nref over A\n  A -> A: drop\nend ref\n@enduml\n",
		},
		"ref between a hidden self-message and its note does not close the window": {
			Src:      "@startuml\nA -> A: drop\nref over B : ./sub.puml\nnote over A: gone\n@enduml\n",
			Keys:     []MessageKey{{"A", "A", "drop"}},
			Expected: "@startuml\nref over B : ./sub.puml\n@enduml\n",
		},
		"note body not affected": {
			Src:      "@startuml\nnote over A\n  A -> A: drop\nend note\n@enduml\n",
			Keys:     []MessageKey{{"A", "A", "drop"}},
			Expected: "@startuml\nnote over A\n  A -> A: drop\nend note\n@enduml\n",
		},
		"self-message: following note dropped": {
			Src:      "@startuml\nA -> A: drop\nnote over A: gone\n@enduml\n",
			Keys:     []MessageKey{{"A", "A", "drop"}},
			Expected: "@startuml\n@enduml\n",
		},
		"self-message: multi-line note dropped": {
			Src:      "@startuml\nA -> A: drop\nnote over A\n  detail\nend note\n@enduml\n",
			Keys:     []MessageKey{{"A", "A", "drop"}},
			Expected: "@startuml\n@enduml\n",
		},
		"bare note left dropped (nil targets)": {
			Src:      "@startuml\nA -> A: drop\nnote left: gone\n@enduml\n",
			Keys:     []MessageKey{{"A", "A", "drop"}},
			Expected: "@startuml\n@enduml\n",
		},
		"note right of A dropped": {
			Src:      "@startuml\nA -> A: drop\nnote right of A: gone\n@enduml\n",
			Keys:     []MessageKey{{"A", "A", "drop"}},
			Expected: "@startuml\n@enduml\n",
		},
		"note left of other participant kept": {
			Src:      "@startuml\nA -> A: drop\nnote left of B: keep\n@enduml\n",
			Keys:     []MessageKey{{"A", "A", "drop"}},
			Expected: "@startuml\nnote left of B: keep\n@enduml\n",
		},
		"run of accompanying notes all dropped": {
			Src:      "@startuml\nA -> A: drop\nnote over A: g1\nnote over A: g2\n@enduml\n",
			Keys:     []MessageKey{{"A", "A", "drop"}},
			Expected: "@startuml\n@enduml\n",
		},
		"multiple trailing notes (multi+single) all dropped": {
			Src:      "@startuml\nA -> A: drop\nnote over A\n  detail\nend note\nnote over A: more\n@enduml\n",
			Keys:     []MessageKey{{"A", "A", "drop"}},
			Expected: "@startuml\n@enduml\n",
		},
		"non-matching note keeps window open; later matching note still dropped": {
			Src:      "@startuml\nA -> A: drop\nnote over B: keep\nnote over A: gone\n@enduml\n",
			Keys:     []MessageKey{{"A", "A", "drop"}},
			Expected: "@startuml\nnote over B: keep\n@enduml\n",
		},
		"delay ... closes the window": {
			Src:      "@startuml\nA -> A: drop\n...\nnote over A: keep\n@enduml\n",
			Keys:     []MessageKey{{"A", "A", "drop"}},
			Expected: "@startuml\n...\nnote over A: keep\n@enduml\n",
		},
		"note over A,B (superset) kept": {
			Src:      "@startuml\nA -> A: drop\nnote over A, B: keep\n@enduml\n",
			Keys:     []MessageKey{{"A", "A", "drop"}},
			Expected: "@startuml\nnote over A, B: keep\n@enduml\n",
		},
		"note after kept message survives": {
			Src:      "@startuml\nA -> B: keep\nnote over A: keep\n@enduml\n",
			Keys:     []MessageKey{{"A", "A", "drop"}},
			Expected: "@startuml\nA -> B: keep\nnote over A: keep\n@enduml\n",
		},
		"note inside block dropped": {
			Src:      "@startuml\nalt c\n  A -> A: drop\n  note over A: gone\nend\n@enduml\n",
			Keys:     []MessageKey{{"A", "A", "drop"}},
			Expected: "@startuml\nalt c\nend\n@enduml\n",
		},
		"rnote accompanying dropped": {
			Src:      "@startuml\nA -> A: drop\nrnote over A: gone\n@enduml\n",
			Keys:     []MessageKey{{"A", "A", "drop"}},
			Expected: "@startuml\n@enduml\n",
		},
		"blank line transparent (still dropped)": {
			Src:      "@startuml\nA -> A: drop\n\nnote over A: gone\n@enduml\n",
			Keys:     []MessageKey{{"A", "A", "drop"}},
			Expected: "@startuml\n\n@enduml\n",
		},
		"non-self hidden message drops its note": {
			Src:      "@startuml\nA -> B: drop\nnote over A: gone\n@enduml\n",
			Keys:     []MessageKey{{"A", "B", "drop"}},
			Expected: "@startuml\n@enduml\n",
		},
		"non-self: note over outsider kept": {
			Src:      "@startuml\nA -> B: drop\nnote over C: keep\n@enduml\n",
			Keys:     []MessageKey{{"A", "B", "drop"}},
			Expected: "@startuml\nnote over C: keep\n@enduml\n",
		},
		"preceding note kept": {
			Src:      "@startuml\nnote over A: keep\nA -> A: drop\n@enduml\n",
			Keys:     []MessageKey{{"A", "A", "drop"}},
			Expected: "@startuml\nnote over A: keep\n@enduml\n",
		},
		"indentation preserved on kept lines": {
			Src:      "@startuml\n  A -> A: drop\n  C -> D: keep\n@enduml\n",
			Keys:     []MessageKey{{"A", "A", "drop"}},
			Expected: "@startuml\n  C -> D: keep\n@enduml\n",
		},
		"empty set returns input unchanged": {
			Src:      "@startuml\nparticipant A\nA -> A: x\n@enduml\n",
			Keys:     nil,
			Expected: "@startuml\nparticipant A\nA -> A: x\n@enduml\n",
		},
		"empty label key": {
			Src:      "@startuml\nA -> A\n@enduml\n",
			Keys:     []MessageKey{{"A", "A", ""}},
			Expected: "@startuml\n@enduml\n",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			d, err := Parse(test.Src)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			got := Hide(d, hidePred(test.Keys...)).String()
			if got != test.Expected {
				t.Errorf("Hide mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}
