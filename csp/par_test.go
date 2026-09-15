package csp

import (
	"fmt"
	"strings"
	"testing"
)

func interleaveAll(ops [][]string) [][]string {
	remaining := 0
	for _, op := range ops {
		remaining += len(op)
	}
	if remaining == 0 {
		return [][]string{{}}
	}
	var out [][]string
	for i, op := range ops {
		if len(op) == 0 {
			continue
		}
		rest := make([][]string, len(ops))
		copy(rest, ops)
		rest[i] = op[1:]
		for _, tail := range interleaveAll(rest) {
			out = append(out, append([]string{op[0]}, tail...))
		}
	}
	return out
}

func altOfTraces(traces [][]string) string {
	var sb strings.Builder
	sb.WriteString("@startuml\n")
	for i, tr := range traces {
		if i == 0 {
			sb.WriteString("alt case0\n")
		} else {
			sb.WriteString(fmt.Sprintf("else case%d\n", i))
		}
		for _, line := range tr {
			sb.WriteString(line + "\n")
		}
	}
	sb.WriteString("end\n@enduml\n")
	return sb.String()
}

func labelLines(traces [][]string) [][]string {
	out := make([][]string, len(traces))
	for i, tr := range traces {
		lines := make([]string, len(tr))
		for j, label := range tr {
			lines[j] = "A -> B: " + label
		}
		out[i] = lines
	}
	return out
}

func TestParIsFullInterleaving(t *testing.T) {
	tests := map[string]struct {
		Par string
		Ops [][]string
	}{
		"two operands of one message": {
			Par: "@startuml\npar\nA -> B: x\nelse\nA -> B: p\nend\n@enduml\n",
			Ops: [][]string{{"x"}, {"p"}},
		},
		"two operands of two messages": {
			Par: "@startuml\npar\nA -> B: x\nA -> B: y\nelse\nA -> B: p\nA -> B: q\nend\n@enduml\n",
			Ops: [][]string{{"x", "y"}, {"p", "q"}},
		},
		"operands of different lengths": {
			Par: "@startuml\npar\nA -> B: x\nA -> B: y\nelse\nA -> B: p\nend\n@enduml\n",
			Ops: [][]string{{"x", "y"}, {"p"}},
		},
		"three operands": {
			Par: "@startuml\npar\nA -> B: x\nelse\nA -> B: p\nelse\nA -> B: q\nend\n@enduml\n",
			Ops: [][]string{{"x"}, {"p"}, {"q"}},
		},
		"empty operand": {
			Par: "@startuml\npar\nA -> B: x\nelse\nend\n@enduml\n",
			Ops: [][]string{{"x"}, {}},
		},
		"nested par flattens into one interleaving": {
			Par: "@startuml\npar\nA -> B: x\nA -> B: y\nelse\npar\nA -> B: p\nelse\nA -> B: q\nend\nend\n@enduml\n",
			Ops: [][]string{{"x", "y"}, {"p"}, {"q"}},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			alt := altOfTraces(labelLines(interleaveAll(test.Ops)))

			if got := RefinesTraces(mustCompile(t, test.Par), mustCompile(t, alt)); !got.Refines {
				t.Errorf("par accepts a trace outside the interleavings: witness=%v offending=%v", got.Witness, got.Offending)
			}
			if got := RefinesTraces(mustCompile(t, alt), mustCompile(t, test.Par)); !got.Refines {
				t.Errorf("par rejects an interleaving: witness=%v offending=%v", got.Witness, got.Offending)
			}
		})
	}
}

func TestBranchOrderDoesNotMatter(t *testing.T) {
	tests := map[string]struct {
		A string
		B string
	}{
		"alt branches swapped": {
			A: "@startuml\nalt ok\nA -> B: x\nelse ng\nA -> B: y\nend\n@enduml\n",
			B: "@startuml\nalt ng\nA -> B: y\nelse ok\nA -> B: x\nend\n@enduml\n",
		},
		"par branches swapped": {
			A: "@startuml\npar\nA -> B: x\nA -> B: y\nelse\nA -> B: p\nend\n@enduml\n",
			B: "@startuml\npar\nA -> B: p\nelse\nA -> B: x\nA -> B: y\nend\n@enduml\n",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := RefinesTraces(mustCompile(t, test.A), mustCompile(t, test.B)); !got.Refines {
				t.Errorf("A should refine B: witness=%v offending=%v", got.Witness, got.Offending)
			}
			if got := RefinesTraces(mustCompile(t, test.B), mustCompile(t, test.A)); !got.Refines {
				t.Errorf("B should refine A: witness=%v offending=%v", got.Witness, got.Offending)
			}
		})
	}
}
