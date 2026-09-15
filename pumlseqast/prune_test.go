package pumlseqast

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPrune(t *testing.T) {
	tests := map[string]struct {
		Src      string
		Expected string
	}{
		"empty opt block removed entirely": {
			Src:      "@startuml\nopt x\nend\n@enduml\n",
			Expected: "@startuml\n@enduml\n",
		},
		"empty loop block removed entirely": {
			Src:      "@startuml\nloop 3 times\nend\n@enduml\n",
			Expected: "@startuml\n@enduml\n",
		},
		"empty group block removed entirely": {
			Src:      "@startuml\ngroup Execution\nend\n@enduml\n",
			Expected: "@startuml\n@enduml\n",
		},
		"block holding only a ref is kept": {
			Src:      "@startuml\ngroup Execution\n  ref over A, B : ./sub.puml\nend\n@enduml\n",
			Expected: "@startuml\ngroup Execution\n  ref over A, B : ./sub.puml\nend\n@enduml\n",
		},
		"single-branch par holding a ref is unwrapped and dedented": {
			Src:      "@startuml\npar\n    ref over A, B\n    ./sub.puml\n    end ref\nelse\nend par\n@enduml\n",
			Expected: "@startuml\nref over A, B\n./sub.puml\nend ref\n@enduml\n",
		},
		"empty else compartment removed, multi-branch block kept": {
			Src:      "@startuml\nalt c\n  A -> B: x\nelse d\n  C -> D: y\nelse\nend\n@enduml\n",
			Expected: "@startuml\nalt c\n  A -> B: x\nelse d\n  C -> D: y\nend\n@enduml\n",
		},
		"single-branch par unwrapped and dedented (acceptance example)": {
			Src:      "@startuml\npar\n    A -> B: m1\n    note right of A\n    detail\n    end note\nelse\nend par\n@enduml\n",
			Expected: "@startuml\nA -> B: m1\nnote right of A\ndetail\nend note\n@enduml\n",
		},
		"all compartments empty removes whole block": {
			Src:      "@startuml\nalt c\nelse\nend\n@enduml\n",
			Expected: "@startuml\n@enduml\n",
		},
		"empty first compartment promotes else to header": {
			Src:      "@startuml\nalt c1\nelse c2\n  A -> B: x\nelse c3\n  C -> D: y\nend\n@enduml\n",
			Expected: "@startuml\nalt c2\n  A -> B: x\nelse c3\n  C -> D: y\nend\n@enduml\n",
		},
		"single-branch alt without else unwrapped (label lost)": {
			Src:      "@startuml\nalt only\n  A -> B: x\nend\n@enduml\n",
			Expected: "@startuml\nA -> B: x\n@enduml\n",
		},
		"nested empty blocks propagate upward": {
			Src:      "@startuml\nopt outer\n  opt inner\n  end\nend\n@enduml\n",
			Expected: "@startuml\n@enduml\n",
		},
		"note-only compartment treated as empty and removed": {
			Src:      "@startuml\nopt x\n  note over A: hi\nend\n@enduml\n",
			Expected: "@startuml\n@enduml\n",
		},
		"comments and blank lines inside otherwise-empty block removed": {
			Src:      "@startuml\nopt x\n  ' just a comment\n\nend\n@enduml\n",
			Expected: "@startuml\n@enduml\n",
		},
		"non-empty group with message note and autonumber preserved": {
			Src:      "@startuml\ngroup G\n  autonumber\n  A -> B: x\n  note over A: n\nend\n@enduml\n",
			Expected: "@startuml\ngroup G\n  autonumber\n  A -> B: x\n  note over A: n\nend\n@enduml\n",
		},
		"box and top-level autonumber preserved": {
			Src:      "@startuml\nautonumber\nbox \"B\"\nparticipant A\nend box\nA -> A: x\n@enduml\n",
			Expected: "@startuml\nautonumber\nbox \"B\"\nparticipant A\nend box\nA -> A: x\n@enduml\n",
		},
		"unwrap preserves nested block with relative indentation": {
			Src:      "@startuml\npar\n    A -> B: x\n    opt inner\n        C -> D: y\n    end\nelse\nend par\n@enduml\n",
			Expected: "@startuml\nA -> B: x\nopt inner\n    C -> D: y\nend\n@enduml\n",
		},
		"diagram without blocks unchanged": {
			Src:      "@startuml\nA -> B: x\nB -> A: y\n@enduml\n",
			Expected: "@startuml\nA -> B: x\nB -> A: y\n@enduml\n",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			d, err := Parse(test.Src)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			got := Prune(d).String()
			if got != test.Expected {
				t.Errorf("Prune mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}

			d2, err := Parse(got)
			if err != nil {
				t.Fatalf("re-parse error: %v", err)
			}
			if again := Prune(d2).String(); again != got {
				t.Errorf("Prune is not idempotent:\nfirst:  %q\nsecond: %q", got, again)
			}
		})
	}
}
