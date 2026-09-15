package pumlseqc4

import (
	"slices"
	"testing"

	"github.com/Kuniwak/puml-tools/pumlseqast"
	"github.com/google/go-cmp/cmp"
)

func TestExtractSeq(t *testing.T) {
	tests := map[string]struct {
		Input        string
		WantOrder    []participant
		WantMessages []message
	}{
		"declared kinds and labels": {
			Input: "@startuml\nactor customer\nparticipant \"Order Service\" as order\n@enduml\n",
			WantOrder: []participant{
				{ID: "customer", Label: "customer", Kind: KindPerson, Declared: true},
				{ID: "order", Label: "Order Service", Kind: KindSystem, Declared: true},
			},
		},
		"unquoted display name is taken as the label, not the alias": {
			Input: "@startuml\nparticipant Cache as c\nparticipant Long as L #Red\n@enduml\n",
			WantOrder: []participant{
				{ID: "c", Label: "Cache", Kind: KindSystem, Declared: true},
				{ID: "L", Label: "Long", Kind: KindSystem, Declared: true},
			},
		},
		"implicit participants and label normalization": {
			Input: "@startuml\na -> b : <:point_right:> hello\n@enduml\n",
			WantOrder: []participant{
				{ID: "a", Label: "a", Kind: KindSystem},
				{ID: "b", Label: "b", Kind: KindSystem},
			},
			WantMessages: []message{{Src: "a", Dst: "b", Label: "hello"}},
		},
		"ref contributes to neither elements nor Rel": {
			Input: "@startuml\nparticipant a\nparticipant b\nref over a, b : ./sub.puml\n@enduml\n",
			WantOrder: []participant{
				{ID: "a", Label: "a", Kind: KindSystem, Declared: true},
				{ID: "b", Label: "b", Kind: KindSystem, Declared: true},
			},
		},
		"arrows inside a ref body do not become Rel": {
			Input: "@startuml\nparticipant a\nparticipant b\nref over a, b\n  a -> b : inner\nend ref\n@enduml\n",
			WantOrder: []participant{
				{ID: "a", Label: "a", Kind: KindSystem, Declared: true},
				{ID: "b", Label: "b", Kind: KindSystem, Declared: true},
			},
		},
		"self message kept": {
			Input:        "@startuml\na -> a : loop\n@enduml\n",
			WantOrder:    []participant{{ID: "a", Label: "a", Kind: KindSystem}},
			WantMessages: []message{{Src: "a", Dst: "a", Label: "loop"}},
		},
		"messages inside block extracted": {
			Input: "@startuml\nparticipant a\nalt cond\n  a -> b : x\nend\n@enduml\n",
			WantOrder: []participant{
				{ID: "a", Label: "a", Kind: KindSystem, Declared: true},
				{ID: "b", Label: "b", Kind: KindSystem},
			},
			WantMessages: []message{{Src: "a", Dst: "b", Label: "x"}},
		},
		"declared kind kept when later referenced": {
			Input: "@startuml\nactor a\na -> b : hi\n@enduml\n",
			WantOrder: []participant{
				{ID: "a", Label: "a", Kind: KindPerson, Declared: true},
				{ID: "b", Label: "b", Kind: KindSystem},
			},
			WantMessages: []message{{Src: "a", Dst: "b", Label: "hi"}},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			d, err := pumlseqast.Parse(test.Input)
			if err != nil {
				t.Fatal(err)
			}
			gotOrder, gotMsgs := extractSeq(d)
			if !slices.Equal(test.WantOrder, gotOrder) {
				t.Errorf("order mismatch (-want +got):\n%s", cmp.Diff(test.WantOrder, gotOrder, cmp.AllowUnexported(participant{})))
			}
			if !slices.Equal(test.WantMessages, gotMsgs) {
				t.Errorf("messages mismatch (-want +got):\n%s", cmp.Diff(test.WantMessages, gotMsgs, cmp.AllowUnexported(message{})))
			}
		})
	}
}

func TestRefWarnings(t *testing.T) {
	tests := map[string]struct {
		Src      string
		Expected []RefWarning
	}{
		"empty when there is no ref": {
			Src:      "@startuml\nA -> B: x\n@enduml\n",
			Expected: nil,
		},
		"single-line ref": {
			Src:      "@startuml\nA -> B: x\nref over A, B : ./sub.puml\n@enduml\n",
			Expected: []RefWarning{{Line: 3, Targets: []string{"A", "B"}}},
		},
		"multiple refs returned in appearance order": {
			Src: "@startuml\nref over A : ./a.puml\nalt c\n  ref over B, C : ./b.puml\nend\n@enduml\n",
			Expected: []RefWarning{
				{Line: 2, Targets: []string{"A"}},
				{Line: 4, Targets: []string{"B", "C"}},
			},
		},
		"empty for an already-expanded diagram": {
			Src:      "@startuml\nA -> B: x\nB -> A: inner\n@enduml\n",
			Expected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			d, err := pumlseqast.Parse(test.Src)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(test.Expected, RefWarnings(d)); diff != "" {
				t.Errorf("RefWarnings mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRefWarningReason(t *testing.T) {
	tests := map[string]struct {
		Warning  RefWarning
		Expected string
	}{
		"single target": {
			Warning:  RefWarning{Line: 3, Targets: []string{"A"}},
			Expected: "ref over A does not contribute to C4 (pass the referenced diagram as well, or run it through pumlseqexpand)",
		},
		"two targets": {
			Warning:  RefWarning{Line: 3, Targets: []string{"A", "B"}},
			Expected: "ref over A, B does not contribute to C4 (pass the referenced diagram as well, or run it through pumlseqexpand)",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := test.Warning.Reason(); got != test.Expected {
				t.Errorf("Reason mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}

func TestNormalizeLabel(t *testing.T) {
	tests := map[string]struct {
		Input    string
		Expected string
	}{
		"drops a leading emoji shortcode":            {Input: "<:point_right:> hello", Expected: "hello"},
		"keeps a tag-like string at the end of body": {Input: "send <json>", Expected: "send <json>"},
		"keeps a tag inside the body":                {Input: "result <b>OK</b>", Expected: "result <b>OK</b>"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := normalizeLabel(test.Input); got != test.Expected {
				t.Errorf("normalizeLabel(%q) = %q, want %q", test.Input, got, test.Expected)
			}
		})
	}
}

func TestNormalizeBoxLabel(t *testing.T) {
	tests := map[string]struct {
		Input    string
		Expected string
	}{
		"extracts a display name wrapped in decoration":              {Input: "<back:#424242><color:#FFFFFF>  Order Service  </color></back>", Expected: "Order Service"},
		"leaves an undecorated display name as is":                   {Input: "Order Service", Expected: "Order Service"},
		"keeps angle brackets that are not a closing tag at the end": {Input: "send <json>", Expected: "send <json>"},
		"a display name that is only closing tags becomes empty":     {Input: "<color:#FFF></color>", Expected: ""},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := NormalizeBoxLabel(test.Input); got != test.Expected {
				t.Errorf("NormalizeBoxLabel(%q) = %q, want %q", test.Input, got, test.Expected)
			}
		})
	}
}
