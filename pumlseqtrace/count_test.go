package pumlseqtrace

import (
	"testing"

	"github.com/Kuniwak/puml-tools/pumlseqast"
)

func TestCountAtMost(t *testing.T) {
	tests := map[string]struct {
		Src      string
		Limit    int
		Expected int
	}{
		"a linear diagram has one trace":        {Src: "@startuml\nA -> B: x\n@enduml\n", Limit: 10, Expected: 1},
		"alt has one trace per branch":          {Src: "@startuml\nalt\n  A -> B: x\nelse\n  A -> B: y\nend\n@enduml\n", Limit: 10, Expected: 2},
		"par of 2 and 1 events has 3 traces":    {Src: "@startuml\npar\n  A -> B: a\n  A -> B: b\nelse\n  C -> D: c\nend\n@enduml\n", Limit: 10, Expected: 3},
		"the count is capped at the limit":      {Src: "@startuml\npar\n  A -> B: a\n  A -> B: b\nelse\n  C -> D: c\nend\n@enduml\n", Limit: 2, Expected: 2},
		"a limit of zero counts nothing":        {Src: "@startuml\nA -> B: x\n@enduml\n", Limit: 0, Expected: 0},
		"break cuts the traces after the block": {Src: "@startuml\nbreak\n  A -> B: e\nend\nalt\n  A -> B: x\nelse\n  A -> B: y\nend\n@enduml\n", Limit: 10, Expected: 1},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			d, err := pumlseqast.Parse(test.Src)
			if err != nil {
				t.Fatalf("Parse: %s", err)
			}
			if got := CountAtMost(d, Options{LoopUnroll: 1}, test.Limit); got != test.Expected {
				t.Errorf("CountAtMost(_, _, %d) = %d, want %d", test.Limit, got, test.Expected)
			}
		})
	}
}
