package pumlseqast

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestProject(t *testing.T) {
	tests := map[string]struct {
		Src      string
		IDs      []string
		Expected string
	}{
		"keeps only messages touching an id": {
			Src:      "@startuml\nparticipant A\nparticipant B\nparticipant C\nA -> B: req\nA -> C: other\nB -> A: resp\n@enduml\n",
			IDs:      []string{"B"},
			Expected: "@startuml\nparticipant A\nparticipant B\nA -> B: req\nB -> A: resp\n@enduml\n",
		},
		"keeps a self message of a kept id": {
			Src:      "@startuml\nparticipant A\nparticipant B\nA -> A: internal\nB -> B: log\n@enduml\n",
			IDs:      []string{"B"},
			Expected: "@startuml\nparticipant B\nB -> B: log\n@enduml\n",
		},
		"keeps a message between two kept ids": {
			Src:      "@startuml\nparticipant A\nparticipant B\nparticipant C\nA -> A: internal\nB -> C: x\n@enduml\n",
			IDs:      []string{"B", "C"},
			Expected: "@startuml\nparticipant B\nparticipant C\nB -> C: x\n@enduml\n",
		},
		"keeps the declaration of an id without any message": {
			Src:      "@startuml\nparticipant A\nparticipant B\nA -> A: internal\n@enduml\n",
			IDs:      []string{"B"},
			Expected: "@startuml\nparticipant B\n@enduml\n",
		},
		"keeps a declaration referenced by a surviving note": {
			Src:      "@startuml\nparticipant B\nparticipant C\nnote over C: memo\nB -> B: log\n@enduml\n",
			IDs:      []string{"B"},
			Expected: "@startuml\nparticipant B\nparticipant C\nnote over C: memo\nB -> B: log\n@enduml\n",
		},
		"removes the note accompanying a removed message": {
			Src:      "@startuml\nparticipant A\nparticipant B\nA -> A: internal\nnote over A: gone\nB -> B: log\n@enduml\n",
			IDs:      []string{"B"},
			Expected: "@startuml\nparticipant B\nB -> B: log\n@enduml\n",
		},
		"prunes a block emptied by the projection": {
			Src:      "@startuml\nparticipant A\nparticipant B\nopt detail\n  A -> A: internal\nend\nB -> B: log\n@enduml\n",
			IDs:      []string{"B"},
			Expected: "@startuml\nparticipant B\nB -> B: log\n@enduml\n",
		},
		"keeps ref over and the declaration it targets": {
			Src:      "@startuml\nparticipant B\nparticipant D\nB -> B: log\nref over D : ./sub.puml\n@enduml\n",
			IDs:      []string{"B"},
			Expected: "@startuml\nparticipant B\nparticipant D\nB -> B: log\nref over D : ./sub.puml\n@enduml\n",
		},
		"removes the frame of a box left without declarations": {
			Src:      "@startuml\nbox \"Backends\"\nparticipant A\nend box\nparticipant B\nA -> A: internal\nB -> B: log\n@enduml\n",
			IDs:      []string{"B"},
			Expected: "@startuml\nparticipant B\nB -> B: log\n@enduml\n",
		},

		"keeps the activation lines of a removed message": {
			Src:      "@startuml\nparticipant A\nparticipant C\nparticipant B\nactivate C\nA -> C: internal\ndeactivate C\nB -> B: log\n@enduml\n",
			IDs:      []string{"B"},
			Expected: "@startuml\nparticipant C\nparticipant B\nactivate C\ndeactivate C\nB -> B: log\n@enduml\n",
		},
		"an id referenced without a declaration is accepted": {
			Src:      "@startuml\nA -> B: req\nA -> A: internal\n@enduml\n",
			IDs:      []string{"B"},
			Expected: "@startuml\nA -> B: req\n@enduml\n",
		},
		"the projection preserves verbatim arrow styles and indentation": {
			Src:      "@startuml\n  A -->> B : req\n  A ->x C : other\n@enduml\n",
			IDs:      []string{"B"},
			Expected: "@startuml\n  A -->> B : req\n@enduml\n",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			d, err := Parse(test.Src)
			if err != nil {
				t.Fatalf("Parse returned error: %v", err)
			}
			got, err := Project(d, test.IDs)
			if err != nil {
				t.Fatalf("Project returned error: %v", err)
			}
			if diff := cmp.Diff(test.Expected, got.String()); diff != "" {
				t.Errorf("Project mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(test.Src, d.String()); diff != "" {
				t.Errorf("input was mutated (-want +got):\n%s", diff)
			}
		})
	}
}

func TestProjectError(t *testing.T) {
	tests := map[string]struct {
		Src string
		IDs []string
	}{
		"unknown id": {
			Src: "@startuml\nparticipant A\nA -> A: x\n@enduml\n",
			IDs: []string{"Nobody"},
		},
		"unknown id among known ones": {
			Src: "@startuml\nparticipant A\nA -> A: x\n@enduml\n",
			IDs: []string{"A", "Nobody"},
		},
		"no id": {
			Src: "@startuml\nparticipant A\nA -> A: x\n@enduml\n",
			IDs: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			d, err := Parse(test.Src)
			if err != nil {
				t.Fatalf("Parse returned error: %v", err)
			}
			if _, err := Project(d, test.IDs); err == nil {
				t.Errorf("Project returned no error, want an error")
			}
		})
	}
}
