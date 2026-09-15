package pumlseqast

import (
	"testing"

	"github.com/Kuniwak/puml-tools/sets"
	"github.com/google/go-cmp/cmp"
)

func TestDropParticipants(t *testing.T) {
	tests := map[string]struct {
		Src      string
		Drop     []string
		Expected string
	}{
		"drops the matching declaration": {
			Src:      "@startuml\nparticipant A\nparticipant B\nA -> B: x\n@enduml\n",
			Drop:     []string{"B"},
			Expected: "@startuml\nparticipant A\nA -> B: x\n@enduml\n",
		},
		"keeps non matching declarations verbatim": {
			Src:      "@startuml\n  participant   A   #red\nparticipant B\n@enduml\n",
			Drop:     []string{"B"},
			Expected: "@startuml\n  participant   A   #red\n@enduml\n",
		},
		"actor keyword declaration is a target too": {
			Src:      "@startuml\nactor A\nparticipant B\n@enduml\n",
			Drop:     []string{"A"},
			Expected: "@startuml\nparticipant B\n@enduml\n",
		},
		"described declaration is matched by its reference name": {
			Src:      "@startuml\nparticipant \"Cache\" as X\nparticipant B\n@enduml\n",
			Drop:     []string{"X"},
			Expected: "@startuml\nparticipant B\n@enduml\n",
		},
		"drops declarations inside blocks and else clauses": {
			Src:      "@startuml\nalt c\n  participant A\nelse\n  participant B\nend\n@enduml\n",
			Drop:     []string{"A", "B"},
			Expected: "@startuml\nalt c\nelse\nend\n@enduml\n",
		},
		"empty predicate is the identity": {
			Src:      "@startuml\nparticipant A\nA -> A: x\n@enduml\n",
			Drop:     nil,
			Expected: "@startuml\nparticipant A\nA -> A: x\n@enduml\n",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			d, err := Parse(test.Src)
			if err != nil {
				t.Fatalf("Parse returned error: %v", err)
			}
			drop := sets.New(test.Drop...)
			got := DropParticipants(d, func(p *Participant) bool { return drop.Contains(p.Name) })
			if diff := cmp.Diff(test.Expected, got.String()); diff != "" {
				t.Errorf("DropParticipants mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(test.Src, d.String()); diff != "" {
				t.Errorf("input was mutated (-want +got):\n%s", diff)
			}
		})
	}
}
