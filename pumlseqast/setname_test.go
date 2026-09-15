package pumlseqast

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSetName(t *testing.T) {
	tests := map[string]struct {
		Src      string
		Name     string
		Expected string
	}{
		"adds name to nameless startuml": {
			Src:      "@startuml\nA -> B: x\n@enduml\n",
			Name:     "X",
			Expected: "@startuml X\nA -> B: x\n@enduml\n",
		},
		"overwrites existing name": {
			Src:      "@startuml Old\nA -> B: x\n@enduml\n",
			Name:     "Y",
			Expected: "@startuml Y\nA -> B: x\n@enduml\n",
		},
		"empty name produces bare startuml": {
			Src:      "@startuml Old\nA -> B: x\n@enduml\n",
			Name:     "",
			Expected: "@startuml\nA -> B: x\n@enduml\n",
		},
		"preserves missing trailing newline term": {
			Src:      "@startuml",
			Name:     "X",
			Expected: "@startuml X",
		},
		"no start node leaves nodes unchanged": {
			Src:      "A -> B: x\n",
			Name:     "X",
			Expected: "A -> B: x\n",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			d, err := Parse(test.Src)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			got := SetName(d, test.Name)
			if got := got.String(); got != test.Expected {
				t.Errorf("SetName mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}
