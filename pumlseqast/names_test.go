package pumlseqast

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDeclaredNames(t *testing.T) {
	tests := map[string]struct {
		Src      string
		Expected []string
	}{
		"declarations only": {
			Src:      "@startuml\nparticipant A\nactor B\n@enduml\n",
			Expected: []string{"A", "B"},
		},
		"alias is the reference name": {
			Src:      "@startuml\nparticipant \"Cache\" as C\n@enduml\n",
			Expected: []string{"C"},
		},
		"message endpoints are not declarations": {
			Src:      "@startuml\nparticipant A\nA -> B: x\n@enduml\n",
			Expected: []string{"A"},
		},
		"recurses into blocks and else clauses": {
			Src:      "@startuml\nalt c\n  participant A\nelse\n  participant B\nend\n@enduml\n",
			Expected: []string{"A", "B"},
		},
		"no declarations": {
			Src:      "@startuml\nA -> B: x\n@enduml\n",
			Expected: []string{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			d, err := Parse(test.Src)
			if err != nil {
				t.Fatalf("Parse returned error: %v", err)
			}
			if diff := cmp.Diff(test.Expected, DeclaredNames(d.Nodes).Sorted()); diff != "" {
				t.Errorf("DeclaredNames mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestReferencedNames(t *testing.T) {
	tests := map[string]struct {
		Src      string
		Expected []string
	}{
		"message endpoints": {
			Src:      "@startuml\nA -> B: x\n@enduml\n",
			Expected: []string{"A", "B"},
		},
		"declarations are not references": {
			Src:      "@startuml\nparticipant A\n@enduml\n",
			Expected: []string{},
		},
		"note targets": {
			Src:      "@startuml\nnote over C, D: memo\n@enduml\n",
			Expected: []string{"C", "D"},
		},
		"ref targets": {
			Src:      "@startuml\nref over E : ./sub.puml\n@enduml\n",
			Expected: []string{"E"},
		},
		"activation targets": {
			Src:      "@startuml\nactivate F\ndeactivate G\n@enduml\n",
			Expected: []string{"F", "G"},
		},
		"recurses into blocks and else clauses": {
			Src:      "@startuml\nalt c\n  A -> B: x\nelse\n  C -> D: y\nend\n@enduml\n",
			Expected: []string{"A", "B", "C", "D"},
		},
		"no references": {
			Src:      "@startuml\nparticipant A\n@enduml\n",
			Expected: []string{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			d, err := Parse(test.Src)
			if err != nil {
				t.Fatalf("Parse returned error: %v", err)
			}
			if diff := cmp.Diff(test.Expected, ReferencedNames(d.Nodes).Sorted()); diff != "" {
				t.Errorf("ReferencedNames mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
