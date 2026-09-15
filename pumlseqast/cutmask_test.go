package pumlseqast

import (
	"errors"
	"slices"
	"testing"

	"github.com/Kuniwak/puml-tools/pumlseq"
	"github.com/google/go-cmp/cmp"
)

func TestCut(t *testing.T) {
	tests := map[string]struct {
		Src         string
		Label       string
		Name        string
		Expected    string
		ExpectedErr error
	}{
		"worked example keeps frame and referenced participants": {
			Src:      "@startuml\nparticipant A\nparticipant B\nparticipant C\nA -> B: login\ngroup Auth\n  B -> C: verify\n  C -> B: ok\nend\nB -> A: done\n@enduml\n",
			Label:    "Auth",
			Expected: "@startuml\nparticipant B\nparticipant C\ngroup Auth\n  B -> C: verify\n  C -> B: ok\nend\n@enduml\n",
		},
		"participant referenced only by a ref inside the group is kept": {
			Src:      "@startuml\nparticipant A\nparticipant B\nparticipant C\nA -> B: login\ngroup Auth\n  ref over B, C : ./sub.puml\nend\n@enduml\n",
			Label:    "Auth",
			Expected: "@startuml\nparticipant B\nparticipant C\ngroup Auth\n  ref over B, C : ./sub.puml\nend\n@enduml\n",
		},
		"no match returns ErrNoMatch": {
			Src:         "@startuml\ngroup Auth\n  A -> B\nend\n@enduml\n",
			Label:       "Other",
			ExpectedErr: ErrNoMatch,
		},
		"two groups same label drops outside messages": {
			Src:      "@startuml\nparticipant A\nparticipant B\nparticipant C\nA -> C: outside\ngroup G\n  A -> B\nend\ngroup G\n  B -> C\nend\n@enduml\n",
			Label:    "G",
			Expected: "@startuml\nparticipant A\nparticipant B\nparticipant C\ngroup G\n  A -> B\nend\ngroup G\n  B -> C\nend\n@enduml\n",
		},
		"startuml name is dropped when no name given": {
			Src:      "@startuml Seq\nparticipant A\nparticipant B\ngroup G\n  A -> B\nend\n@enduml\n",
			Label:    "G",
			Expected: "@startuml\nparticipant A\nparticipant B\ngroup G\n  A -> B\nend\n@enduml\n",
		},
		"name argument is appended to startuml": {
			Src:      "@startuml Seq\nparticipant A\nparticipant B\ngroup G\n  A -> B\nend\n@enduml\n",
			Label:    "G",
			Name:     "Excerpt",
			Expected: "@startuml Excerpt\nparticipant A\nparticipant B\ngroup G\n  A -> B\nend\n@enduml\n",
		},
		"aliased participant declaration preserved verbatim": {
			Src:      "@startuml\nparticipant \"Auth Service\" as AS\nparticipant B\ngroup G\n  AS -> B\nend\n@enduml\n",
			Label:    "G",
			Expected: "@startuml\nparticipant \"Auth Service\" as AS\nparticipant B\ngroup G\n  AS -> B\nend\n@enduml\n",
		},
		"nested inner match keeps original indentation": {
			Src:      "@startuml\nparticipant A\nparticipant B\nparticipant C\ngroup Outer\n  A -> B\n  group Inner\n    B -> C\n  end\nend\n@enduml\n",
			Label:    "Inner",
			Expected: "@startuml\nparticipant B\nparticipant C\n  group Inner\n    B -> C\n  end\n@enduml\n",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			d, err := Parse(test.Src)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			got, err := Cut(d, pumlseq.Exact(test.Label), test.Name)
			if test.ExpectedErr != nil {
				if !errors.Is(err, test.ExpectedErr) {
					t.Errorf("error mismatch: got %v, want %v", err, test.ExpectedErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := got.String(); got != test.Expected {
				t.Errorf("Cut mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}

func TestMask(t *testing.T) {
	tests := map[string]struct {
		Src         string
		Label       string
		Name        string
		Expected    string
		ExpectedErr error
	}{
		"worked example drops participant used only inside": {
			Src:      "@startuml\nparticipant A\nparticipant B\nparticipant C\nA -> B: login\ngroup Auth\n  B -> C: verify\n  C -> B: ok\nend\nB -> A: done\n@enduml\n",
			Label:    "Auth",
			Expected: "@startuml\nparticipant A\nparticipant B\nA -> B: login\nB -> A: done\n@enduml\n",
		},
		"participant referenced only by a ref outside the group is kept": {
			Src:      "@startuml\nparticipant A\nparticipant B\nparticipant C\nref over A, C : ./sub.puml\ngroup Auth\n  B -> C: verify\nend\n@enduml\n",
			Label:    "Auth",
			Expected: "@startuml\nparticipant A\nparticipant C\nref over A, C : ./sub.puml\n@enduml\n",
		},
		"startuml name is dropped when no name given": {
			Src:      "@startuml Seq\nparticipant A\nparticipant B\nA -> B: x\ngroup G\n  A -> B\nend\n@enduml\n",
			Label:    "G",
			Expected: "@startuml\nparticipant A\nparticipant B\nA -> B: x\n@enduml\n",
		},
		"name argument is appended to startuml": {
			Src:      "@startuml Seq\nparticipant A\nparticipant B\nA -> B: x\ngroup G\n  A -> B\nend\n@enduml\n",
			Label:    "G",
			Name:     "Masked",
			Expected: "@startuml Masked\nparticipant A\nparticipant B\nA -> B: x\n@enduml\n",
		},
		"no match returns ErrNoMatch": {
			Src:         "@startuml\ngroup Auth\n  A -> B\nend\n@enduml\n",
			Label:       "Other",
			ExpectedErr: ErrNoMatch,
		},
		"participant referenced inside and outside stays": {
			Src:      "@startuml\nparticipant A\nparticipant B\nparticipant C\nA -> C: keep\ngroup G\n  A -> B\n  B -> C\nend\n@enduml\n",
			Label:    "G",
			Expected: "@startuml\nparticipant A\nparticipant C\nA -> C: keep\n@enduml\n",
		},
		"nested group removed with parent": {
			Src:      "@startuml\nparticipant A\nparticipant B\nparticipant C\nA -> B: x\ngroup Outer\n  B -> C\n  group Inner\n    C -> A\n  end\nend\n@enduml\n",
			Label:    "Outer",
			Expected: "@startuml\nparticipant A\nparticipant B\nA -> B: x\n@enduml\n",
		},
		"two groups same label both removed": {
			Src:      "@startuml\nparticipant A\nparticipant B\nparticipant C\nparticipant D\ngroup G\n  A -> B\nend\nC -> D: keep\ngroup G\n  C -> A\nend\n@enduml\n",
			Label:    "G",
			Expected: "@startuml\nparticipant C\nparticipant D\nC -> D: keep\n@enduml\n",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			d, err := Parse(test.Src)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			got, err := Mask(d, pumlseq.Exact(test.Label), test.Name)
			if test.ExpectedErr != nil {
				if !errors.Is(err, test.ExpectedErr) {
					t.Errorf("error mismatch: got %v, want %v", err, test.ExpectedErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := got.String(); got != test.Expected {
				t.Errorf("Mask mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}

func TestOutermostGroups(t *testing.T) {
	tests := map[string]struct {
		Src      string
		Expected []string
	}{
		"all messages wrapped in groups": {
			Src:      "@startuml\nparticipant A\nparticipant B\ngroup Login\n  A -> B\nend\ngroup Order\n  B -> A\nend\n@enduml\n",
			Expected: []string{"Login", "Order"},
		},
		"no group returns nothing": {
			Src:      "@startuml\nA -> B\n@enduml\n",
			Expected: nil,
		},
		"nested group is not returned by itself": {
			Src:      "@startuml\ngroup Outer\n  group Inner\n    A -> B\n  end\nend\n@enduml\n",
			Expected: []string{"Outer"},
		},
		"group inside alt is outermost": {
			Src:      "@startuml\nalt cond\n  group Inner\n    A -> B\n  end\nelse\n  group Other\n    A -> B\n  end\nend\n@enduml\n",
			Expected: []string{"Inner", "Other"},
		},
		"duplicated labels are returned as many times as they occur": {
			Src:      "@startuml\ngroup G\n  A -> B\nend\ngroup G\n  B -> A\nend\n@enduml\n",
			Expected: []string{"G", "G"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			d, err := Parse(test.Src)
			if err != nil {
				t.Fatalf("Parse returned error: %v", err)
			}

			var got []string
			for _, b := range OutermostGroups(d) {
				got = append(got, b.Label)
			}
			if !slices.Equal(test.Expected, got) {
				t.Errorf("OutermostGroups labels mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}
