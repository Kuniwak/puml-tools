package pumlseqbreak

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFind(t *testing.T) {
	tests := map[string]struct {
		Src      string
		Expected []Break
	}{
		"clean chain": {
			Src:      "@startuml\na -> b\nb -> c\n@enduml\n",
			Expected: nil,
		},
		"empty": {
			Src:      "",
			Expected: nil,
		},
		"header only": {
			Src:      "@startuml\n@enduml\n",
			Expected: nil,
		},
		"sender without trigger": {
			Src:      "@startuml\na -> b\nc -> b\n@enduml\n",
			Expected: []Break{MessageBreak(3, "c")},
		},
		"fan-out adjacent is ok": {
			Src:      "@startuml\na -> b\na -> c\n@enduml\n",
			Expected: nil,
		},
		"markovian adjacency breaks both": {
			Src:      "@startuml\na -> b\nx -> y\na -> c\n@enduml\n",
			Expected: []Break{MessageBreak(3, "x"), MessageBreak(4, "a")},
		},
		"ref establishes context for its targets": {
			Src:      "@startuml\na -> b\nref over b, c : ./sub.puml\nc -> d\n@enduml\n",
			Expected: nil,
		},
		"ref without a preceding trigger is a break": {
			Src:      "@startuml\na -> b\nref over c, d : ./sub.puml\n@enduml\n",
			Expected: []Break{RefBreak(3, []string{"c", "d"})},
		},
		"ref at the head of the diagram is not a break": {
			Src:      "@startuml\nref over a, b : ./sub.puml\na -> b\n@enduml\n",
			Expected: nil,
		},
		"multiline ref establishes context and its body is opaque": {
			Src:      "@startuml\na -> b\nref over b, c\n  x -> y\nend ref\nc -> d\n@enduml\n",
			Expected: nil,
		},
		"message after a ref whose targets do not include the sender is a break": {
			Src:      "@startuml\na -> b\nref over b, c : ./sub.puml\nx -> y\n@enduml\n",
			Expected: []Break{MessageBreak(4, "x")},
		},
		"a ref without targets does not disrupt context": {
			Src:      "@startuml\na -> b\nref\n  sub.puml\nend ref\nb -> c\n@enduml\n",
			Expected: nil,
		},
		"a ref without targets is itself never reported as a break": {
			Src:      "@startuml\na -> b\nref\n  sub.puml\nend ref\nx -> y\n@enduml\n",
			Expected: []Break{MessageBreak(6, "x")},
		},
		"a ref inside a par branch also establishes exit context": {
			Src:      "@startuml\na -> b\npar\n  ref over c, d : ./sub.puml\nelse\n  b -> e\nend\nd -> f\n@enduml\n",
			Expected: []Break{RefBreak(4, []string{"c", "d"})},
		},
		"delay resets causality": {
			Src:      "@startuml\na -> b\n...\nc -> b\n@enduml\n",
			Expected: nil,
		},
		"delay with text resets causality": {
			Src:      "@startuml\na -> b\n...5 minutes later...\nc -> b\n@enduml\n",
			Expected: nil,
		},
		"divider does not reset": {
			Src:      "@startuml\na -> b\n== Phase 2 ==\nc -> b\n@enduml\n",
			Expected: []Break{MessageBreak(4, "c")},
		},
		"newpage does not reset": {
			Src:      "@startuml\na -> b\nnewpage\nc -> b\n@enduml\n",
			Expected: []Break{MessageBreak(4, "c")},
		},
		"declarations and activation are skipped": {
			Src:      "@startuml\na -> b\nparticipant Z\nactivate b\nb -> c\n@enduml\n",
			Expected: nil,
		},
		"line comment is not a message": {
			Src:      "@startuml\na -> b\n'a -> b\nb -> c\n@enduml\n",
			Expected: nil,
		},
		"single line note is skipped": {
			Src:      "@startuml\na -> b\nnote right of b: hello\nb -> c\n@enduml\n",
			Expected: nil,
		},
		"multi line note body is not scanned": {
			Src:      "@startuml\na -> b\nnote over a\nx -> y\nend note\nb -> c\n@enduml\n",
			Expected: nil,
		},
		"multi line block comment body is not scanned": {
			Src:      "@startuml\na -> b\n/'\nx -> y\n'/\nb -> c\n@enduml\n",
			Expected: nil,
		},
		"alt branch first message chains from entry": {
			Src:      "@startuml\na -> b\nalt\nb -> c\nend\nc -> d\n@enduml\n",
			Expected: nil,
		},
		"alt branch first message without trigger breaks": {
			Src:      "@startuml\na -> b\nalt\nd -> e\nend\n@enduml\n",
			Expected: []Break{MessageBreak(4, "d")},
		},
		"resend after optional block chains from entry": {
			Src:      "@startuml\na -> b\nopt\nb -> c\nend\na -> z\n@enduml\n",
			Expected: nil,
		},
		"else branch resends from entry": {
			Src:      "@startuml\na -> b\nalt\nb -> c\nelse\na -> z\nend\n@enduml\n",
			Expected: nil,
		},
		"else branch first message without trigger breaks": {
			Src:      "@startuml\na -> b\nalt\nb -> c\nelse\nd -> e\nend\n@enduml\n",
			Expected: []Break{MessageBreak(6, "d")},
		},
		"parallel block always runs so entry is dropped": {
			Src:      "@startuml\na -> b\npar\nb -> c\nend\na -> z\n@enduml\n",
			Expected: []Break{MessageBreak(6, "a")},
		},
		"delay in a branch resets after the block": {
			Src:      "@startuml\na -> b\nalt\n...\nelse\nb -> e\nend\nz -> w\n@enduml\n",
			Expected: nil,
		},
		"nested alt in loop is clean": {
			Src:      "@startuml\na -> b\nloop\nalt\nb -> c\nelse\nb -> d\nend\nend\n@enduml\n",
			Expected: nil,
		},
		"nested inner branch without trigger breaks": {
			Src:      "@startuml\na -> b\nloop\nalt\nb -> c\nelse\nx -> y\nend\nend\n@enduml\n",
			Expected: []Break{MessageBreak(7, "x")},
		},
		"missing end does not panic": {
			Src:      "@startuml\na -> b\nalt\nb -> c\n@enduml\n",
			Expected: nil,
		},
		"stray else is ignored": {
			Src:      "@startuml\na -> b\nelse\nb -> c\n@enduml\n",
			Expected: nil,
		},
		"stray end is ignored": {
			Src:      "@startuml\na -> b\nend\nb -> c\n@enduml\n",
			Expected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := Find([]byte(test.Src))

			if diff := cmp.Diff(test.Expected, got); diff != "" {
				t.Errorf("Find mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBreakReason(t *testing.T) {
	tests := map[string]struct {
		Break    Break
		Expected string
	}{
		"plain sender":  {Break: MessageBreak(3, "c"), Expected: "sender c has no preceding trigger"},
		"quoted sender": {Break: MessageBreak(3, "Svc One"), Expected: "sender Svc One has no preceding trigger"},
		"ref subject":   {Break: RefBreak(3, []string{"c", "d"}), Expected: "ref over c, d has no preceding trigger"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := test.Break.Reason()
			if got != test.Expected {
				t.Errorf("Reason mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}

func TestBreakReasonPanicsOnUnknownKind(t *testing.T) {
	tests := map[string]struct {
		Kind Kind
	}{
		"zero value":    {Kind: ""},
		"unknown value": {Kind: "unknown"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("Reason with kind %q should panic, but returned", test.Kind)
				}
			}()
			_ = Break{Line: 1, Kind: test.Kind, Lifelines: []string{"a"}}.Reason()
		})
	}
}
