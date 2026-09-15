package pumlseqast

import (
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParticipants(t *testing.T) {
	tests := map[string]struct {
		Src      string
		Expected []string
	}{
		"declarations and references": {
			Src:      "@startuml\nparticipant A\nparticipant B\nA -> B: x\nB -> C: y\n@enduml\n",
			Expected: []string{"A", "B", "C"},
		},
		"references only": {
			Src:      "@startuml\nA -> B: x\n@enduml\n",
			Expected: []string{"A", "B"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			d, err := Parse(test.Src)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			if got := Participants(d); !slices.Equal(test.Expected, got) {
				t.Errorf("Participants mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}

func TestFoldAll(t *testing.T) {
	tests := map[string]struct {
		Src      string
		Phi      Folding
		Expected string
	}{
		"single group keeps internal as self message": {
			Src:      "@startuml\nparticipant A\nparticipant B\nparticipant C\nA -> B: req\nB -> C: fwd\nC -> B: resp\nB -> A: resp\n@enduml\n",
			Phi:      Folding{"B": "BE", "C": "BE"},
			Expected: "@startuml\nparticipant A\nparticipant BE\nA -> BE: req\nBE -> BE: fwd\nBE -> BE: resp\nBE -> A: resp\n@enduml\n",
		},
		"identity mapping is skipped": {
			Src:      "@startuml\nparticipant A\nparticipant B\nA -> B: x\n@enduml\n",
			Phi:      Folding{"A": "A", "B": "B"},
			Expected: "@startuml\nparticipant A\nparticipant B\nA -> B: x\n@enduml\n",
		},
		"single element rename": {
			Src:      "@startuml\nparticipant A\nparticipant B\nA -> B: x\n@enduml\n",
			Phi:      Folding{"B": "BE"},
			Expected: "@startuml\nparticipant A\nparticipant BE\nA -> BE: x\n@enduml\n",
		},
		"two groups": {
			Src:      "@startuml\nparticipant A\nparticipant B\nparticipant C\nA -> B: req\nB -> C: fwd\n@enduml\n",
			Phi:      Folding{"A": "FE", "B": "BE", "C": "BE"},
			Expected: "@startuml\nparticipant FE\nparticipant BE\nFE -> BE: req\nBE -> BE: fwd\n@enduml\n",
		},
		"merge into an existing abstract participant": {
			Src:      "@startuml\nparticipant A\nparticipant B\nparticipant C\nA -> B: req\nB -> C: fwd\n@enduml\n",
			Phi:      Folding{"A": "A", "B": "A", "C": "A"},
			Expected: "@startuml\nparticipant A\nA -> A: req\nA -> A: fwd\n@enduml\n",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			d, err := Parse(test.Src)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			got, err := FoldAll(d, test.Phi)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := got.String(); got != test.Expected {
				t.Errorf("FoldAll mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}

func TestValidateFolding(t *testing.T) {
	tests := map[string]struct {
		Abstract    string
		Detail      string
		Phi         Folding
		ExpectedErr bool
	}{
		"all detail participants map into abstract": {
			Abstract:    "@startuml\nparticipant A\nparticipant BE\nA -> BE: req\n@enduml\n",
			Detail:      "@startuml\nparticipant A\nparticipant B\nparticipant C\nA -> B: req\nB -> C: fwd\n@enduml\n",
			Phi:         Folding{"B": "BE", "C": "BE"},
			ExpectedErr: false,
		},
		"unmapped participant is the identity and must exist in abstract": {
			Abstract:    "@startuml\nparticipant A\nparticipant B\nA -> B: req\n@enduml\n",
			Detail:      "@startuml\nparticipant A\nparticipant B\nA -> B: req\n@enduml\n",
			Phi:         Folding{},
			ExpectedErr: false,
		},
		"image is not a participant of abstract": {
			Abstract:    "@startuml\nparticipant A\nparticipant BE\nA -> BE: req\n@enduml\n",
			Detail:      "@startuml\nparticipant A\nparticipant B\nA -> B: req\n@enduml\n",
			Phi:         Folding{"B": "FE"},
			ExpectedErr: true,
		},
		"unmapped participant missing from abstract": {
			Abstract:    "@startuml\nparticipant A\nparticipant BE\nA -> BE: req\n@enduml\n",
			Detail:      "@startuml\nparticipant A\nparticipant B\nA -> B: req\n@enduml\n",
			Phi:         Folding{},
			ExpectedErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			abstract, err := Parse(test.Abstract)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			detail, err := Parse(test.Detail)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			if got := ValidateFolding(test.Phi, abstract, detail); (got != nil) != test.ExpectedErr {
				t.Errorf("ValidateFolding returned %v, want error = %v", got, test.ExpectedErr)
			}
		})
	}
}
