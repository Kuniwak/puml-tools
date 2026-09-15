package pumlseqast

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFold(t *testing.T) {
	tests := map[string]struct {
		Src      string
		IDs      []string
		NewID    string
		NewDesc  string
		Expected string
	}{
		"two participants merged with desc": {
			Src:      "@startuml\nparticipant A\nparticipant B\nparticipant C\nA -> B: req\nB -> C: fwd\nC -> A: resp\n@enduml\n",
			IDs:      []string{"B", "C"},
			NewID:    "BC",
			NewDesc:  "Backend",
			Expected: "@startuml\nparticipant A\nparticipant \"Backend\" as BC\nA -> BC: req\nBC -> BC: fwd\nBC -> A: resp\n@enduml\n",
		},
		"empty desc produces bare declaration": {
			Src:      "@startuml\nparticipant A\nparticipant B\nA -> B\n@enduml\n",
			IDs:      []string{"B"},
			NewID:    "BB",
			NewDesc:  "",
			Expected: "@startuml\nparticipant A\nparticipant BB\nA -> BB\n@enduml\n",
		},
		"touchpoint is reference before declaration": {
			Src:      "@startuml\nA -> B: req\nparticipant B\n@enduml\n",
			IDs:      []string{"B"},
			NewID:    "BB",
			Expected: "@startuml\nparticipant BB\nA -> BB: req\n@enduml\n",
		},
		"reverse arrow renamed inserts declaration at first reference": {
			Src:      "@startuml\nparticipant A\nparticipant B\nb <- a\n@enduml\n",
			IDs:      []string{"a"},
			NewID:    "z",
			Expected: "@startuml\nparticipant A\nparticipant B\nparticipant z\nb <- z\n@enduml\n",
		},
		"quoted endpoint renamed": {
			Src:      "@startuml\n\"Svc One\" -> B\n@enduml\n",
			IDs:      []string{"Svc One"},
			NewID:    "S",
			Expected: "@startuml\nparticipant S\nS -> B\n@enduml\n",
		},
		"reference inside alt and else renamed": {
			Src:      "@startuml\nparticipant A\nparticipant B\nalt ok\n  A -> B\nelse ng\n  B -> A\nend\n@enduml\n",
			IDs:      []string{"B"},
			NewID:    "BB",
			Expected: "@startuml\nparticipant A\nparticipant BB\nalt ok\n  A -> BB\nelse ng\n  BB -> A\nend\n@enduml\n",
		},
		"declaration inside block is dropped and hoisted to top level": {
			Src:      "@startuml\nparticipant A\ngroup G\n  A -> B: x\nend\n@enduml\n",
			IDs:      []string{"B"},
			NewID:    "BB",
			Expected: "@startuml\nparticipant A\nparticipant BB\ngroup G\n  A -> BB: x\nend\n@enduml\n",
		},
		"note targets deduped after rename": {
			Src:      "@startuml\nparticipant A\nparticipant B\nnote over A, B: hi\n@enduml\n",
			IDs:      []string{"A", "B"},
			NewID:    "AB",
			Expected: "@startuml\nparticipant AB\nnote over AB: hi\n@enduml\n",
		},
		"note partial rename keeps others": {
			Src:      "@startuml\nparticipant A\nparticipant B\nparticipant C\nnote over A, C: hi\n@enduml\n",
			IDs:      []string{"A"},
			NewID:    "AA",
			Expected: "@startuml\nparticipant AA\nparticipant B\nparticipant C\nnote over AA, C: hi\n@enduml\n",
		},
		"ref targets deduped after rename": {
			Src:      "@startuml\nparticipant A\nparticipant B\nref over A, B: ./sub.puml\n@enduml\n",
			IDs:      []string{"A", "B"},
			NewID:    "AB",
			Expected: "@startuml\nparticipant AB\nref over AB: ./sub.puml\n@enduml\n",
		},
		"ref partial rename keeps others": {
			Src:      "@startuml\nparticipant A\nparticipant B\nparticipant C\nref over A, C: ./sub.puml\n@enduml\n",
			IDs:      []string{"A"},
			NewID:    "AA",
			Expected: "@startuml\nparticipant AA\nparticipant B\nparticipant C\nref over AA, C: ./sub.puml\n@enduml\n",
		},
		"multiline ref targets renamed": {
			Src:      "@startuml\nparticipant A\nparticipant B\nref over A, B\n  ./sub.puml\nend ref\n@enduml\n",
			IDs:      []string{"A", "B"},
			NewID:    "AB",
			Expected: "@startuml\nparticipant AB\nref over AB\n  ./sub.puml\nend ref\n@enduml\n",
		},
		"participant referenced only by ref is a touchpoint": {
			Src:      "@startuml\nref over A, B: ./sub.puml\n@enduml\n",
			IDs:      []string{"B"},
			NewID:    "BB",
			Expected: "@startuml\nparticipant BB\nref over A, BB: ./sub.puml\n@enduml\n",
		},
		"note rename keeps the space before the colon": {
			Src:      "@startuml\nparticipant A\nparticipant B\nnote over A, B : hi\n@enduml\n",
			IDs:      []string{"A", "B"},
			NewID:    "AB",
			Expected: "@startuml\nparticipant AB\nnote over AB : hi\n@enduml\n",
		},
		"ref rename keeps the space before the colon": {
			Src:      "@startuml\nparticipant A\nparticipant B\nref over A, B : ./sub.puml\n@enduml\n",
			IDs:      []string{"A", "B"},
			NewID:    "AB",
			Expected: "@startuml\nparticipant AB\nref over AB : ./sub.puml\n@enduml\n",
		},
		"activation target renamed": {
			Src:      "@startuml\nparticipant A\nparticipant B\nactivate B\nA -> B\ndeactivate B\n@enduml\n",
			IDs:      []string{"B"},
			NewID:    "BB",
			Expected: "@startuml\nparticipant A\nparticipant BB\nactivate BB\nA -> BB\ndeactivate BB\n@enduml\n",
		},
		"activation rename preserves spacing (lossless)": {
			Src:      "@startuml\nparticipant A\nparticipant B\nactivate   B\nA -> B\n@enduml\n",
			IDs:      []string{"B"},
			NewID:    "BB",
			Expected: "@startuml\nparticipant A\nparticipant BB\nactivate   BB\nA -> BB\n@enduml\n",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			d, err := Parse(test.Src)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			got, err := Fold(d, test.IDs, test.NewID, test.NewDesc)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := got.String(); got != test.Expected {
				t.Errorf("Fold mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}

func TestFoldErrors(t *testing.T) {
	tests := map[string]struct {
		Src     string
		IDs     []string
		NewID   string
		WantErr string
	}{
		"missing participant": {
			Src:     "@startuml\nparticipant A\nA -> A\n@enduml\n",
			IDs:     []string{"X"},
			NewID:   "Y",
			WantErr: "participant not found: X",
		},
		"new id collides with existing participant": {
			Src:     "@startuml\nparticipant A\nparticipant B\nparticipant C\nA -> B\n@enduml\n",
			IDs:     []string{"A"},
			NewID:   "C",
			WantErr: "conflicts with an existing participant",
		},
		"no ids given": {
			Src:     "@startuml\nA -> B\n@enduml\n",
			IDs:     nil,
			NewID:   "Z",
			WantErr: "at least one",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			d, err := Parse(test.Src)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			_, err = Fold(d, test.IDs, test.NewID, "")
			if err == nil {
				t.Fatalf("want error containing %q, got nil", test.WantErr)
			}
			if !strings.Contains(err.Error(), test.WantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), test.WantErr)
			}
		})
	}
}
