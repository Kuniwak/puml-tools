package csp

import (
	"slices"
	"testing"

	"github.com/Kuniwak/puml-tools/pumlseqast"
	"github.com/google/go-cmp/cmp"
)

func mustCompile(t *testing.T, src string) *LTS {
	t.Helper()
	l, err := CompileSource(src)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}
	return l
}

func TestRefinesTraces(t *testing.T) {
	tests := map[string]struct {
		Spec      string
		Impl      string
		Refines   bool
		Witness   []pumlseqast.MessageKey
		Offending pumlseqast.MessageKey
	}{
		"identical single message": {
			Spec:    "@startuml\nA -> B: x\n@enduml\n",
			Impl:    "@startuml\nA -> B: x\n@enduml\n",
			Refines: true,
		},
		"impl emits extra message not allowed by spec": {
			Spec:      "@startuml\nA -> B: x\n@enduml\n",
			Impl:      "@startuml\nA -> B: x\nA -> B: y\n@enduml\n",
			Refines:   false,
			Witness:   []pumlseqast.MessageKey{{Sender: "A", Receiver: "B", Label: "x"}},
			Offending: pumlseqast.MessageKey{Sender: "A", Receiver: "B", Label: "y"},
		},
		"impl is a prefix of spec": {
			Spec:    "@startuml\nA -> B: x\nA -> B: y\n@enduml\n",
			Impl:    "@startuml\nA -> B: x\n@enduml\n",
			Refines: true,
		},
		"label mismatch is not a refinement": {
			Spec:      "@startuml\nA -> B: x\n@enduml\n",
			Impl:      "@startuml\nA -> B: z\n@enduml\n",
			Refines:   false,
			Offending: pumlseqast.MessageKey{Sender: "A", Receiver: "B", Label: "z"},
		},
		"impl implements only one alt branch": {
			Spec:    "@startuml\nalt ok\nA -> B: x\nelse ng\nA -> B: y\nend\n@enduml\n",
			Impl:    "@startuml\nA -> B: x\n@enduml\n",
			Refines: true,
		},
		"impl chooses a branch not in spec alt": {
			Spec:      "@startuml\nalt ok\nA -> B: x\nelse ng\nA -> B: y\nend\n@enduml\n",
			Impl:      "@startuml\nA -> B: z\n@enduml\n",
			Refines:   false,
			Offending: pumlseqast.MessageKey{Sender: "A", Receiver: "B", Label: "z"},
		},
		"alt refines the same alt": {
			Spec:    "@startuml\nalt ok\nA -> B: x\nelse ng\nA -> B: y\nend\n@enduml\n",
			Impl:    "@startuml\nalt ok\nA -> B: x\nelse ng\nA -> B: y\nend\n@enduml\n",
			Refines: true,
		},
		"opt taken refines opt": {
			Spec:    "@startuml\nopt c\nA -> B: x\nend\n@enduml\n",
			Impl:    "@startuml\nA -> B: x\n@enduml\n",
			Refines: true,
		},
		"opt skipped refines opt": {
			Spec:    "@startuml\nopt c\nA -> B: x\nend\n@enduml\n",
			Impl:    "@startuml\n@enduml\n",
			Refines: true,
		},
		"loop refines the same loop": {
			Spec:    "@startuml\nloop\nA -> B: x\nend\n@enduml\n",
			Impl:    "@startuml\nloop\nA -> B: x\nend\n@enduml\n",
			Refines: true,
		},
		"one iteration refines loop": {
			Spec:    "@startuml\nloop\nA -> B: x\nend\n@enduml\n",
			Impl:    "@startuml\nA -> B: x\n@enduml\n",
			Refines: true,
		},
		"loop body not in spec is not a refinement": {
			Spec:      "@startuml\nloop\nA -> B: x\nend\n@enduml\n",
			Impl:      "@startuml\nA -> B: y\n@enduml\n",
			Refines:   false,
			Offending: pumlseqast.MessageKey{Sender: "A", Receiver: "B", Label: "y"},
		},
		"one interleaving refines par": {
			Spec:    "@startuml\npar\nA -> B: x\nelse\nA -> B: y\nend\n@enduml\n",
			Impl:    "@startuml\nA -> B: x\nA -> B: y\n@enduml\n",
			Refines: true,
		},
		"the other interleaving refines par": {
			Spec:    "@startuml\npar\nA -> B: x\nelse\nA -> B: y\nend\n@enduml\n",
			Impl:    "@startuml\nA -> B: y\nA -> B: x\n@enduml\n",
			Refines: true,
		},
		"repeating one operand violates par": {
			Spec:      "@startuml\npar\nA -> B: x\nelse\nA -> B: y\nend\n@enduml\n",
			Impl:      "@startuml\nA -> B: x\nA -> B: x\n@enduml\n",
			Refines:   false,
			Witness:   []pumlseqast.MessageKey{{Sender: "A", Receiver: "B", Label: "x"}},
			Offending: pumlseqast.MessageKey{Sender: "A", Receiver: "B", Label: "x"},
		},
		"par refines the same par": {
			Spec:    "@startuml\npar\nA -> B: x\nelse\nA -> B: y\nend\n@enduml\n",
			Impl:    "@startuml\npar\nA -> B: x\nelse\nA -> B: y\nend\n@enduml\n",
			Refines: true,
		},
		"group is transparent": {
			Spec:    "@startuml\nA -> B: x\n@enduml\n",
			Impl:    "@startuml\ngroup label\nA -> B: x\nend\n@enduml\n",
			Refines: true,
		},
		"critical is transparent": {
			Spec:    "@startuml\nA -> B: x\n@enduml\n",
			Impl:    "@startuml\ncritical\nA -> B: x\nend\n@enduml\n",
			Refines: true,
		},
		"break is transparent": {
			Spec:    "@startuml\nA -> B: x\n@enduml\n",
			Impl:    "@startuml\nbreak\nA -> B: x\nend\n@enduml\n",
			Refines: true,
		},
		"group body is not skipped": {
			Spec:      "@startuml\nA -> B: x\n@enduml\n",
			Impl:      "@startuml\ngroup label\nA -> B: y\nend\n@enduml\n",
			Refines:   false,
			Offending: pumlseqast.MessageKey{Sender: "A", Receiver: "B", Label: "y"},
		},
		"empty impl refines anything": {
			Spec:    "@startuml\nA -> B: x\n@enduml\n",
			Impl:    "@startuml\n@enduml\n",
			Refines: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := RefinesTraces(mustCompile(t, test.Impl), mustCompile(t, test.Spec))
			if got.Refines != test.Refines {
				t.Fatalf("Refines = %v, want %v (witness=%v offending=%v)", got.Refines, test.Refines, got.Witness, got.Offending)
			}
			if test.Refines {
				return
			}
			if !slices.Equal(test.Witness, got.Witness) {
				t.Errorf("witness mismatch (-want +got):\n%s", cmp.Diff(test.Witness, got.Witness))
			}
			if got.Offending != test.Offending {
				t.Errorf("offending = %v, want %v", got.Offending, test.Offending)
			}
		})
	}
}
