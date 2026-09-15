package pumlseqgroupsplit

import (
	"errors"
	"github.com/Kuniwak/puml-tools/outfiles"
	"slices"
	"testing"

	"github.com/Kuniwak/puml-tools/pumlseqast"
	"github.com/google/go-cmp/cmp"
)

func TestSplit(t *testing.T) {
	tests := map[string]struct {
		Src         string
		Basename    string
		Expected    []outfiles.File
		ExpectedErr error
	}{
		"worked example splits each group into its own file": {
			Src:      "@startuml\nparticipant A\nparticipant B\nparticipant C\ngroup Login\n  A -> B: login\nend\ngroup Verify\n  B -> C: verify\nend\n@enduml\n",
			Basename: "seq",
			Expected: []outfiles.File{
				{Name: "seq_1_Login.puml", Content: "@startuml\nparticipant A\nparticipant B\ngroup Login\n  A -> B: login\nend\n@enduml\n"},
				{Name: "seq_2_Verify.puml", Content: "@startuml\nparticipant B\nparticipant C\ngroup Verify\n  B -> C: verify\nend\n@enduml\n"},
			},
		},
		"no group is an error": {
			Src:         "@startuml\nA -> B\n@enduml\n",
			Basename:    "seq",
			ExpectedErr: ErrNoGroup,
		},
		"duplicated titles are distinguished by index": {
			Src:      "@startuml\nparticipant A\nparticipant B\ngroup G\n  A -> B\nend\ngroup G\n  B -> A\nend\n@enduml\n",
			Basename: "seq",
			Expected: []outfiles.File{
				{Name: "seq_1_G.puml", Content: "@startuml\nparticipant A\nparticipant B\ngroup G\n  A -> B\nend\n@enduml\n"},
				{Name: "seq_2_G.puml", Content: "@startuml\nparticipant A\nparticipant B\ngroup G\n  B -> A\nend\n@enduml\n"},
			},
		},
		"characters unusable in a file name are replaced with underscore": {
			Src:      "@startuml\nparticipant A\nparticipant B\ngroup Order / Confirm?\n  A -> B\nend\n@enduml\n",
			Basename: "seq",
			Expected: []outfiles.File{
				{Name: "seq_1_Order___Confirm_.puml", Content: "@startuml\nparticipant A\nparticipant B\ngroup Order / Confirm?\n  A -> B\nend\n@enduml\n"},
			},
		},
		"empty title omits the title part": {
			Src:      "@startuml\nparticipant A\nparticipant B\ngroup\n  A -> B\nend\n@enduml\n",
			Basename: "seq",
			Expected: []outfiles.File{
				{Name: "seq_1.puml", Content: "@startuml\nparticipant A\nparticipant B\ngroup\n  A -> B\nend\n@enduml\n"},
			},
		},
		"a message outside a group is an error": {
			Src:         "@startuml\nparticipant A\nparticipant B\nA -> B: outside\ngroup G\n  A -> B\nend\n@enduml\n",
			Basename:    "seq",
			ExpectedErr: ErrMessageOutsideGroup,
		},
		"a message outside a group inside an alt is an error": {
			Src:         "@startuml\nparticipant A\nparticipant B\nalt cond\n  A -> B: outside\nend\ngroup G\n  A -> B\nend\n@enduml\n",
			Basename:    "seq",
			ExpectedErr: ErrMessageOutsideGroup,
		},
		"basename with a path separator is an error": {
			Src:         "@startuml\nparticipant A\nparticipant B\ngroup G\n  A -> B\nend\n@enduml\n",
			Basename:    "../escaped",
			ExpectedErr: outfiles.ErrBasenameHasPathSeparator,
		},
		"nested group stays inside the outer group file": {
			Src:      "@startuml\nparticipant A\nparticipant B\nparticipant C\ngroup Outer\n  A -> B\n  group Inner\n    B -> C\n  end\nend\n@enduml\n",
			Basename: "seq",
			Expected: []outfiles.File{
				{Name: "seq_1_Outer.puml", Content: "@startuml\nparticipant A\nparticipant B\nparticipant C\ngroup Outer\n  A -> B\n  group Inner\n    B -> C\n  end\nend\n@enduml\n"},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			d, err := pumlseqast.Parse(test.Src)
			if err != nil {
				t.Fatalf("Parse returned error: %v", err)
			}

			got, err := Split(d, test.Basename)
			if !errors.Is(err, test.ExpectedErr) {
				t.Fatalf("Split returned error %v, want %v", err, test.ExpectedErr)
			}
			if !slices.Equal(test.Expected, got) {
				t.Errorf("Split mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}
