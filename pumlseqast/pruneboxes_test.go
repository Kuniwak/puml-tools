package pumlseqast

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPruneEmptyBoxes(t *testing.T) {
	tests := map[string]struct {
		Src      string
		Expected string
	}{
		"removes the frame of an emptied box": {
			Src:      "@startuml\nbox \"Group\"\nend box\nA -> A: x\n@enduml\n",
			Expected: "@startuml\nA -> A: x\n@enduml\n",
		},
		"keeps a box that still encloses a declaration": {
			Src:      "@startuml\nbox \"Group\"\nparticipant A\nend box\n@enduml\n",
			Expected: "@startuml\nbox \"Group\"\nparticipant A\nend box\n@enduml\n",
		},
		"keeps a box that encloses a declaration inside a block": {
			Src:      "@startuml\nbox \"Group\"\nalt c\nparticipant A\nend\nend box\n@enduml\n",
			Expected: "@startuml\nbox \"Group\"\nalt c\nparticipant A\nend\nend box\n@enduml\n",
		},
		"diagram without box is unchanged": {
			Src:      "@startuml\nparticipant A\nA -> A: x\n@enduml\n",
			Expected: "@startuml\nparticipant A\nA -> A: x\n@enduml\n",
		},
		"removes only the empty one of several boxes": {
			Src:      "@startuml\nbox \"Kept\"\nparticipant A\nend box\nbox \"Emptied\"\nend box\nA -> A: x\n@enduml\n",
			Expected: "@startuml\nbox \"Kept\"\nparticipant A\nend box\nA -> A: x\n@enduml\n",
		},
		"keeps comments and blank lines inside a removed box": {
			Src:      "@startuml\nbox \"Group\"\n' memo\n\nend box\n@enduml\n",
			Expected: "@startuml\n' memo\n\n@enduml\n",
		},
		"box with a color specification": {
			Src:      "@startuml\nbox \"Group\" #LightBlue\nend box\n@enduml\n",
			Expected: "@startuml\n@enduml\n",
		},
		"box without a label": {
			Src:      "@startuml\nbox\nend box\n@enduml\n",
			Expected: "@startuml\n@enduml\n",
		},
		"endbox spelling": {
			Src:      "@startuml\nbox \"Group\"\nendbox\n@enduml\n",
			Expected: "@startuml\n@enduml\n",
		},
		"CRLF line terminators are preserved": {
			Src:      "@startuml\r\nbox \"Group\"\r\nend box\r\n@enduml\r\n",
			Expected: "@startuml\r\n@enduml\r\n",
		},
		"no trailing newline on the last line": {
			Src:      "@startuml\nbox \"Group\"\nend box",
			Expected: "@startuml\n",
		},
		"unclosed box is left alone": {
			Src:      "@startuml\nbox \"Group\"\n@enduml\n",
			Expected: "@startuml\nbox \"Group\"\n@enduml\n",
		},
		"indented box frame is removed": {
			Src:      "@startuml\n  box \"Group\"\n  end box\n@enduml\n",
			Expected: "@startuml\n@enduml\n",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			d, err := Parse(test.Src)
			if err != nil {
				t.Fatalf("Parse returned error: %v", err)
			}
			got := PruneEmptyBoxes(d)
			if diff := cmp.Diff(test.Expected, got.String()); diff != "" {
				t.Errorf("PruneEmptyBoxes mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(test.Src, d.String()); diff != "" {
				t.Errorf("input was mutated (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(test.Expected, PruneEmptyBoxes(got).String()); diff != "" {
				t.Errorf("not idempotent (-want +got):\n%s", diff)
			}
		})
	}
}
