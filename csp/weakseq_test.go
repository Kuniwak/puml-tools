package csp

import (
	"strings"
	"testing"
)

type testMsg struct {
	Sender   string
	Receiver string
	Label    string
}

func (m testMsg) line() string { return m.Sender + " -> " + m.Receiver + ": " + m.Label }

func (m testMsg) shares(o testMsg) bool {
	return m.Sender == o.Sender || m.Sender == o.Receiver ||
		m.Receiver == o.Sender || m.Receiver == o.Receiver
}

func linearExtensions(msgs []testMsg) [][]string {
	var out [][]string
	done := make([]bool, len(msgs))
	var rec func(acc []string)
	rec = func(acc []string) {
		if len(acc) == len(msgs) {
			out = append(out, append([]string(nil), acc...))
			return
		}
		for i, m := range msgs {
			if done[i] {
				continue
			}
			ready := true
			for j := 0; j < i; j++ {
				if !done[j] && msgs[j].shares(m) {
					ready = false
					break
				}
			}
			if !ready {
				continue
			}
			done[i] = true
			rec(append(acc, m.line()))
			done[i] = false
		}
	}
	rec(nil)
	return out
}

func seqOf(msgs []testMsg) string {
	var sb strings.Builder
	sb.WriteString("@startuml\n")
	for _, m := range msgs {
		sb.WriteString(m.line() + "\n")
	}
	sb.WriteString("@enduml\n")
	return sb.String()
}

func TestSeqIsWeakSequencing(t *testing.T) {
	tests := map[string]struct {
		Msgs []testMsg
	}{
		"disjoint participants are unordered": {
			Msgs: []testMsg{{"A", "B", "x"}, {"C", "D", "y"}},
		},
		"a shared participant orders them": {
			Msgs: []testMsg{{"A", "B", "x"}, {"B", "C", "y"}},
		},
		"same sender orders them": {
			Msgs: []testMsg{{"A", "B", "x"}, {"A", "C", "y"}},
		},
		"a chain is totally ordered": {
			Msgs: []testMsg{{"A", "B", "x"}, {"B", "C", "y"}, {"C", "D", "z"}},
		},
		"two independent messages joined by a third": {
			Msgs: []testMsg{{"A", "B", "x"}, {"C", "D", "y"}, {"B", "C", "z"}},
		},
		"three mutually disjoint messages": {
			Msgs: []testMsg{{"A", "B", "x"}, {"C", "D", "y"}, {"E", "F", "z"}},
		},
		"four messages in two independent chains": {
			Msgs: []testMsg{{"A", "B", "x"}, {"C", "D", "y"}, {"B", "A", "z"}, {"D", "C", "w"}},
		},
		"a late message joins two disjoint chains": {
			Msgs: []testMsg{{"A", "B", "x"}, {"C", "D", "y"}, {"E", "F", "z"}, {"B", "E", "w"}},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			diagram := seqOf(test.Msgs)
			alt := altOfTraces(linearExtensions(test.Msgs))

			if got := RefinesTraces(mustCompile(t, diagram), mustCompile(t, alt)); !got.Refines {
				t.Errorf("the diagram accepts a trace that is not a linear extension: witness=%v offending=%v", got.Witness, got.Offending)
			}
			if got := RefinesTraces(mustCompile(t, alt), mustCompile(t, diagram)); !got.Refines {
				t.Errorf("the diagram rejects a linear extension: witness=%v offending=%v", got.Witness, got.Offending)
			}
		})
	}
}

func TestWeakSequencingIsPerLifeline(t *testing.T) {

	group := "@startuml\ngroup g\nA -> B: x\nC -> D: w\nend\nC -> E: z\n@enduml\n"

	tests := map[string]struct {
		Spec            string
		Impl            string
		ExpectedRefines bool
	}{
		"a lifeline not shared with the block has no ordering with what follows": {
			Spec:            group,
			Impl:            "@startuml\nC -> D: w\nC -> E: z\nA -> B: x\n@enduml\n",
			ExpectedRefines: true,
		},
		"matches the same message sequence with the block removed": {
			Spec:            "@startuml\nA -> B: x\nC -> D: w\nC -> E: z\n@enduml\n",
			Impl:            "@startuml\nC -> D: w\nC -> E: z\nA -> B: x\n@enduml\n",
			ExpectedRefines: true,
		},
		"preserves declaration order on a shared lifeline": {
			Spec:            group,
			Impl:            "@startuml\nC -> E: z\nC -> D: w\nA -> B: x\n@enduml\n",
			ExpectedRefines: false,
		},
		"a lifeline in an unchosen branch does not block what follows": {

			Spec:            "@startuml\nalt cond\nA -> B: x\nelse\nC -> D: w\nend\nC -> E: z\n@enduml\n",
			Impl:            "@startuml\nC -> E: z\nA -> B: x\n@enduml\n",
			ExpectedRefines: true,
		},
		"preserves order on the chosen branch's lifeline": {
			Spec:            "@startuml\nalt cond\nA -> B: x\nelse\nC -> D: w\nend\nC -> E: z\n@enduml\n",
			Impl:            "@startuml\nC -> E: z\nC -> D: w\n@enduml\n",
			ExpectedRefines: false,
		},
		"a lifeline outside a loop has no ordering with what follows": {
			Spec:            "@startuml\nloop 3\nA -> B: x\nend\nC -> D: z\n@enduml\n",
			Impl:            "@startuml\nC -> D: z\nA -> B: x\n@enduml\n",
			ExpectedRefines: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := RefinesTraces(mustCompile(t, test.Impl), mustCompile(t, test.Spec))
			if got.Refines != test.ExpectedRefines {
				t.Errorf("Refines = %v, want %v (witness=%v offending=%v)", got.Refines, test.ExpectedRefines, got.Witness, got.Offending)
			}
		})
	}
}

func groupOf(msgs []testMsg, from, to int) string {
	var sb strings.Builder
	sb.WriteString("@startuml\n")
	for i, m := range msgs {
		if i == from {
			sb.WriteString("group g\n")
		}
		if i == to {
			sb.WriteString("end\n")
		}
		sb.WriteString(m.line() + "\n")
	}
	if to == len(msgs) {
		sb.WriteString("end\n")
	}
	sb.WriteString("@enduml\n")
	return sb.String()
}

func TestTransparentBlockKeepsTraces(t *testing.T) {
	tests := map[string]struct {
		Msgs []testMsg
		From int
		To   int
	}{
		"wraps the first 2 messages (partially shared with what follows)": {
			Msgs: []testMsg{{"A", "B", "x"}, {"C", "D", "w"}, {"C", "E", "z"}},
			From: 0, To: 2,
		},
		"wraps the middle 2 messages": {
			Msgs: []testMsg{{"A", "B", "x"}, {"C", "D", "w"}, {"E", "F", "v"}, {"C", "E", "z"}},
			From: 1, To: 3,
		},
		"wraps through to the end": {
			Msgs: []testMsg{{"A", "B", "x"}, {"C", "D", "w"}, {"C", "E", "z"}},
			From: 1, To: 3,
		},
		"wraps everything": {
			Msgs: []testMsg{{"A", "B", "x"}, {"C", "D", "w"}, {"C", "E", "z"}},
			From: 0, To: 3,
		},
		"wraps part of a chain": {
			Msgs: []testMsg{{"A", "B", "x"}, {"B", "C", "y"}, {"C", "D", "z"}},
			From: 0, To: 2,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			flat := mustCompile(t, seqOf(test.Msgs))
			grouped := mustCompile(t, groupOf(test.Msgs, test.From, test.To))

			if got := RefinesTraces(grouped, flat); !got.Refines {
				t.Errorf("wrapping in a group added traces: witness=%v offending=%v", got.Witness, got.Offending)
			}
			if got := RefinesTraces(flat, grouped); !got.Refines {
				t.Errorf("wrapping in a group removed traces: witness=%v offending=%v", got.Witness, got.Offending)
			}
		})
	}
}
