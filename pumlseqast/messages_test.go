package pumlseqast

import (
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMessageLine(t *testing.T) {

	src := "@startuml\nA -> B\nnote over X\n  Y -> Z\nend note\nC -> D\nalt cond\n  E -> F\nend\n@enduml\n"
	d, err := Parse(src)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	var got []int
	for _, m := range Messages(d) {
		got = append(got, m.Line)
	}
	if !slices.Equal([]int{2, 6, 8}, got) {
		t.Errorf("message lines mismatch (-want +got):\n%s", cmp.Diff([]int{2, 6, 8}, got))
	}
}

func TestMessagesAndKey(t *testing.T) {
	tests := map[string]struct {
		Src      string
		Expected []MessageKey
	}{
		"top level and direction resolved": {
			Src:      "@startuml\nA -> B: req\nb <- a\n@enduml\n",
			Expected: []MessageKey{{Sender: "A", Receiver: "B", Label: "req"}, {Sender: "a", Receiver: "b", Label: ""}},
		},
		"self message and no label": {
			Src:      "@startuml\nX -> X: loop\nA -> A\n@enduml\n",
			Expected: []MessageKey{{Sender: "X", Receiver: "X", Label: "loop"}, {Sender: "A", Receiver: "A", Label: ""}},
		},
		"recurses into block then and else clauses in order": {
			Src:      "@startuml\nA -> B: top\nalt c\n  C -> D: then\nelse\n  E -> F: els\nend\nG -> H: tail\n@enduml\n",
			Expected: []MessageKey{{"A", "B", "top"}, {"C", "D", "then"}, {"E", "F", "els"}, {"G", "H", "tail"}},
		},
		"quoted endpoints unquoted in key": {
			Src:      "@startuml\n\"Svc One\" -> B: hi\n@enduml\n",
			Expected: []MessageKey{{Sender: "Svc One", Receiver: "B", Label: "hi"}},
		},
		"note body is not a message": {
			Src:      "@startuml\nnote over A\n  X -> Y: inside note\nend note\nA -> B: real\n@enduml\n",
			Expected: []MessageKey{{Sender: "A", Receiver: "B", Label: "real"}},
		},
		"commented messages are not messages": {
			Src:      "@startuml\n'X -> Y: line comment\n/'\nP -> Q: block comment\n'/\nA -> B: real\n@enduml\n",
			Expected: []MessageKey{{Sender: "A", Receiver: "B", Label: "real"}},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			d, err := Parse(test.Src)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			msgs := Messages(d)
			got := make([]MessageKey, len(msgs))
			for i, m := range msgs {
				got[i] = m.Key()
			}
			if !slices.Equal(test.Expected, got) {
				t.Errorf("Messages/Key mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}

func TestMessageKeyString(t *testing.T) {
	tests := map[string]struct {
		Key      MessageKey
		Expected string
	}{
		"with label":    {Key: MessageKey{Sender: "A", Receiver: "B", Label: "x"}, Expected: "A -> B: x"},
		"without label": {Key: MessageKey{Sender: "A", Receiver: "B"}, Expected: "A -> B"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := test.Key.String(); got != test.Expected {
				t.Errorf("String() = %q, want %q", got, test.Expected)
			}
		})
	}
}

func TestCompareMessageKey(t *testing.T) {
	tests := map[string]struct {
		A        MessageKey
		B        MessageKey
		Expected int
	}{
		"equal": {
			A:        MessageKey{Sender: "A", Receiver: "B", Label: "x"},
			B:        MessageKey{Sender: "A", Receiver: "B", Label: "x"},
			Expected: 0,
		},
		"sender decides first": {
			A:        MessageKey{Sender: "A", Receiver: "Z", Label: "z"},
			B:        MessageKey{Sender: "B", Receiver: "A", Label: "a"},
			Expected: -1,
		},
		"receiver decides when senders are equal": {
			A:        MessageKey{Sender: "A", Receiver: "C", Label: "a"},
			B:        MessageKey{Sender: "A", Receiver: "B", Label: "z"},
			Expected: 1,
		},
		"label decides last": {
			A:        MessageKey{Sender: "A", Receiver: "B", Label: "x"},
			B:        MessageKey{Sender: "A", Receiver: "B", Label: "y"},
			Expected: -1,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := CompareMessageKey(test.A, test.B)
			if (got < 0) != (test.Expected < 0) || (got == 0) != (test.Expected == 0) || (got > 0) != (test.Expected > 0) {
				t.Errorf("CompareMessageKey = %d, want the same sign as %d", got, test.Expected)
			}
		})
	}
}

func TestMessagesIn(t *testing.T) {
	tests := map[string]struct {
		Src      string
		Expected []string
	}{
		"top level messages in order": {
			Src:      "@startuml\nA -> B: x\nB -> C: y\n@enduml\n",
			Expected: []string{"A -> B: x", "B -> C: y"},
		},
		"messages inside blocks and else clauses": {
			Src:      "@startuml\nA -> B: x\nalt ok\nB -> C: y\nelse ng\nB -> D: z\nend\n@enduml\n",
			Expected: []string{"A -> B: x", "B -> C: y", "B -> D: z"},
		},
		"declarations and notes are not messages": {
			Src:      "@startuml\nparticipant A\nnote over A: hi\n@enduml\n",
			Expected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			d, err := Parse(test.Src)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			var got []string
			for _, m := range MessagesIn(d.Nodes) {
				got = append(got, m.Key().String())
			}
			if !slices.Equal(test.Expected, got) {
				t.Errorf("MessagesIn = %v, want %v", got, test.Expected)
			}
		})
	}
}

func TestMessagesOutsideGroups(t *testing.T) {
	tests := map[string]struct {
		Src      string
		Expected []int
	}{
		"empty when every message is enclosed in a group": {
			Src:      "@startuml\ngroup G\n  A -> B\nend\ngroup H\n  B -> A\nend\n@enduml\n",
			Expected: nil,
		},
		"a message before a group is outside the group": {
			Src:      "@startuml\nA -> B\ngroup G\n  B -> C\nend\n@enduml\n",
			Expected: []int{2},
		},
		"a message after a group is outside the group": {
			Src:      "@startuml\ngroup G\n  A -> B\nend\nB -> A\n@enduml\n",
			Expected: []int{5},
		},
		"a message between two groups is outside the group": {
			Src:      "@startuml\ngroup G\n  A -> B\nend\nA -> C\ngroup H\n  B -> A\nend\n@enduml\n",
			Expected: []int{5},
		},
		"a message inside a nested group is inside the group": {
			Src:      "@startuml\ngroup Outer\n  group Inner\n    A -> B\n  end\nend\n@enduml\n",
			Expected: nil,
		},
		"a message inside an alt not enclosed in a group is outside the group": {
			Src:      "@startuml\nalt cond\n  A -> B\nelse\n  A -> C\nend\n@enduml\n",
			Expected: []int{3, 5},
		},
		"a message inside a group nested in an alt is inside the group": {
			Src:      "@startuml\nalt cond\n  group G\n    A -> B\n  end\nend\n@enduml\n",
			Expected: nil,
		},
		"every message is outside the group when there is no group": {
			Src:      "@startuml\nA -> B\nB -> A\n@enduml\n",
			Expected: []int{2, 3},
		},
		"empty when there are no messages": {
			Src:      "@startuml\nparticipant A\n@enduml\n",
			Expected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			d, err := Parse(test.Src)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			var got []int
			for _, m := range MessagesOutsideGroups(d) {
				got = append(got, m.Line)
			}
			if !slices.Equal(test.Expected, got) {
				t.Errorf("MessagesOutsideGroups lines mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}
