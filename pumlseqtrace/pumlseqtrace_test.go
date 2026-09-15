package pumlseqtrace

import (
	"testing"

	"github.com/Kuniwak/puml-tools/pumlseqast"
	"github.com/google/go-cmp/cmp"
)

func TestTraces(t *testing.T) {
	tests := map[string]struct {
		Src      string
		Options  Options
		Expected [][]string
	}{
		"a linear diagram has exactly one trace": {
			Src:      "@startuml\nA -> B: x\nB -> A: y\n@enduml\n",
			Expected: [][]string{{"x", "y"}},
		},
		"a diagram without messages has the empty trace": {
			Src:      "@startuml\nparticipant A\n@enduml\n",
			Expected: [][]string{{}},
		},
		"alt yields one trace per branch in order": {
			Src:      "@startuml\nA -> B: x\nalt ok\n  A -> B: y\nelse ng\n  A -> B: z\nend\nA -> B: w\n@enduml\n",
			Expected: [][]string{{"x", "y", "w"}, {"x", "z", "w"}},
		},
		"opt yields the branch and the skip": {
			Src:      "@startuml\nopt maybe\n  A -> B: y\nend\nA -> B: w\n@enduml\n",
			Expected: [][]string{{"y", "w"}, {"w"}},
		},
		"break truncates the trace there": {
			Src:      "@startuml\nA -> B: x\nbreak error\n  A -> B: e\nend\nA -> B: w\n@enduml\n",
			Expected: [][]string{{"x", "e"}},
		},
		"break inside alt truncates only the branch that reaches it": {
			Src:      "@startuml\nalt ok\n  A -> B: y\nelse ng\n  break error\n    A -> B: e\n  end\nend\nA -> B: w\n@enduml\n",
			Expected: [][]string{{"y", "w"}, {"e"}},
		},
		"loop is unrolled once by default": {
			Src:      "@startuml\nloop 3 times\n  A -> B: y\nend\nA -> B: w\n@enduml\n",
			Options:  Options{LoopUnroll: 1},
			Expected: [][]string{{"y", "w"}},
		},
		"loop unroll 2 yields one trace per number of turns": {
			Src:      "@startuml\nloop 3 times\n  A -> B: y\nend\nA -> B: w\n@enduml\n",
			Options:  Options{LoopUnroll: 2},
			Expected: [][]string{{"y", "w"}, {"y", "y", "w"}},
		},
		"loop unroll 0 skips the body": {
			Src:      "@startuml\nloop 3 times\n  A -> B: y\nend\nA -> B: w\n@enduml\n",
			Options:  Options{LoopUnroll: 0},
			Expected: [][]string{{"w"}},
		},
		"par interleaves the operands": {
			Src:      "@startuml\npar\n  A -> B: x\nelse\n  C -> D: y\nend\n@enduml\n",
			Expected: [][]string{{"x", "y"}, {"y", "x"}},
		},
		"par keeps the order inside each operand": {
			Src:      "@startuml\npar\n  A -> B: a\n  A -> B: b\nelse\n  C -> D: c\nend\n@enduml\n",
			Expected: [][]string{{"a", "b", "c"}, {"a", "c", "b"}, {"c", "a", "b"}},
		},
		"par interleaves each combination of branch choices": {
			Src:      "@startuml\npar\n  alt ok\n    A -> B: x\n  else ng\n    A -> B: z\n  end\nelse\n  C -> D: y\nend\n@enduml\n",
			Expected: [][]string{{"x", "y"}, {"y", "x"}, {"z", "y"}, {"y", "z"}},
		},
		"break inside par stops the trace at the break, not after the join": {
			Src:      "@startuml\npar\n  break error\n    A -> B: e\n  end\nelse\n  C -> D: y\nend\nA -> B: w\n@enduml\n",
			Expected: [][]string{{"e"}, {"y", "e"}},
		},
		"break inside par cuts the rest of the other operand": {
			Src:      "@startuml\npar\n  break error\n    A -> B: e\n  end\nelse\n  C -> D: y1\n  C -> D: y2\nend\n@enduml\n",
			Expected: [][]string{{"e"}, {"y1", "e"}, {"y1", "y2", "e"}},
		},
		"an empty break inside par stops the trace at once": {
			Src:      "@startuml\npar\n  break error\n  end\nelse\n  C -> D: y\nend\n@enduml\n",
			Expected: [][]string{{}},
		},
		"break inside a loop stops before the next turn": {
			Src:      "@startuml\nloop\n  A -> B: x\n  break error\n    A -> B: e\n  end\nend\nA -> B: w\n@enduml\n",
			Options:  Options{LoopUnroll: 3},
			Expected: [][]string{{"x", "e"}},
		},
		"nested loops are unrolled independently": {
			Src:     "@startuml\nloop outer\n  loop inner\n    A -> B: i\n  end\nend\n@enduml\n",
			Options: Options{LoopUnroll: 2},

			Expected: [][]string{{"i"}, {"i", "i"}, {"i", "i", "i"}, {"i", "i"}, {"i", "i", "i"}, {"i", "i", "i", "i"}},
		},
		"a loop inside par is unrolled before interleaving": {
			Src:      "@startuml\npar\n  loop\n    A -> B: a\n  end\nelse\n  C -> D: c\nend\n@enduml\n",
			Options:  Options{LoopUnroll: 2},
			Expected: [][]string{{"a", "c"}, {"c", "a"}, {"a", "a", "c"}, {"a", "c", "a"}, {"c", "a", "a"}},
		},
		"group and critical are transparent": {
			Src:      "@startuml\ngroup g\n  A -> B: x\nend\ncritical c\n  A -> B: y\nend\n@enduml\n",
			Expected: [][]string{{"x", "y"}},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			d, err := pumlseqast.Parse(test.Src)
			if err != nil {
				t.Fatalf("Parse: %s", err)
			}
			got := labels(Traces(d, test.Options))
			if diff := cmp.Diff(test.Expected, got); diff != "" {
				t.Errorf("Traces mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func labels(traces [][]pumlseqast.MessageKey) [][]string {
	out := make([][]string, len(traces))
	for i, tr := range traces {
		out[i] = []string{}
		for _, k := range tr {
			out[i] = append(out[i], k.Label)
		}
	}
	return out
}

func TestTracesPanicsOnRef(t *testing.T) {
	d, err := pumlseqast.Parse("@startuml\nref over A, B: other.puml\n@enduml\n")
	if err != nil {
		t.Fatalf("Parse: %s", err)
	}

	defer func() {
		if recover() == nil {
			t.Errorf("Traces should panic on an unexpanded ref, but did not")
		}
	}()
	Traces(d, Options{LoopUnroll: 1})
}

func TestWalkStopsWhenYieldReturnsFalse(t *testing.T) {

	src := "@startuml\npar\n  A -> B: a1\n  A -> B: a2\n  A -> B: a3\n  A -> B: a4\n  A -> B: a5\nelse\n  C -> D: b1\n  C -> D: b2\n  C -> D: b3\n  C -> D: b4\n  C -> D: b5\nend\n@enduml\n"
	d, err := pumlseqast.Parse(src)
	if err != nil {
		t.Fatalf("Parse: %s", err)
	}

	seen := 0
	Walk(d, Options{LoopUnroll: 1}, func([]pumlseqast.MessageKey) bool {
		seen++
		return seen < 3
	})
	if seen != 3 {
		t.Errorf("Walk yielded %d traces after the 3rd was refused, want 3", seen)
	}
}
