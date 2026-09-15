package c4ast

import (
	"os"
	"reflect"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
	tests := map[string]struct {
		Input    string
		Expected *Diagram
	}{
		"single person element": {
			Input: "@startuml\nPerson(u, \"User\")\n@enduml\n",
			Expected: &Diagram{
				Order:      []ElementID{"u"},
				Elements:   map[ElementID]Element{"u": {ID: "u", Kind: "Person", Label: "User"}},
				Boundaries: map[ElementID]Boundary{},
			},
		},
		"diagram name": {
			Input: "@startuml CTX\nSystem(a, \"A\")\n@enduml\n",
			Expected: &Diagram{
				Name:       "CTX",
				Order:      []ElementID{"a"},
				Elements:   map[ElementID]Element{"a": {ID: "a", Kind: "System", Label: "A"}},
				Boundaries: map[ElementID]Boundary{},
			},
		},
		"external with description": {
			Input: "@startuml\nSystem_Ext(x, \"X\", \"D\")\n@enduml\n",
			Expected: &Diagram{
				Order:      []ElementID{"x"},
				Elements:   map[ElementID]Element{"x": {ID: "x", Kind: "System", External: true, Label: "X", Descr: "D"}},
				Boundaries: map[ElementID]Boundary{},
			},
		},
		"element base kinds preserved": {
			Input: "@startuml\nSystemDb(a, \"A\")\nSystemQueue(b, \"B\")\nContainer(c, \"C\")\nComponentDb(d, \"D\")\n@enduml\n",
			Expected: &Diagram{
				Order: []ElementID{"a", "b", "c", "d"},
				Elements: map[ElementID]Element{
					"a": {ID: "a", Kind: "SystemDb", Label: "A"},
					"b": {ID: "b", Kind: "SystemQueue", Label: "B"},
					"c": {ID: "c", Kind: "Container", Label: "C"},
					"d": {ID: "d", Kind: "ComponentDb", Label: "D"},
				},
				Boundaries: map[ElementID]Boundary{},
			},
		},
		"rel family": {
			Input: "@startuml\nSystem(a, \"A\")\nSystem(b, \"B\")\n" +
				"Rel(a, b, \"r1\")\nRel_Back(a, b, \"r2\")\nBiRel(a, b, \"r3\")\nRel_D(a, b, \"r4\")\nRel(a, b)\n@enduml\n",
			Expected: &Diagram{
				Order: []ElementID{"a", "b"},
				Elements: map[ElementID]Element{
					"a": {ID: "a", Kind: "System", Label: "A"},
					"b": {ID: "b", Kind: "System", Label: "B"},
				},
				Boundaries: map[ElementID]Boundary{},
				Rels: []Rel{
					{From: "a", To: "b", Label: "r1"},
					{From: "b", To: "a", Label: "r2"},
					{From: "a", To: "b", Label: "r3"},
					{From: "b", To: "a", Label: "r3"},
					{From: "a", To: "b", Label: "r4"},
					{From: "a", To: "b", Label: ""},
				},
			},
		},
		"boundary nesting": {
			Input: "@startuml\n" +
				"Enterprise_Boundary(ent, \"Ent\") {\n" +
				"  System(s1, \"S1\")\n" +
				"  System_Boundary(sb, \"SB\") {\n" +
				"    System(s2, \"S2\")\n" +
				"  }\n" +
				"}\n" +
				"System(top, \"Top\")\n" +
				"@enduml\n",
			Expected: &Diagram{
				Order: []ElementID{"s1", "s2", "top"},
				Elements: map[ElementID]Element{
					"s1":  {ID: "s1", Kind: "System", Label: "S1", Boundary: "ent"},
					"s2":  {ID: "s2", Kind: "System", Label: "S2", Boundary: "sb"},
					"top": {ID: "top", Kind: "System", Label: "Top"},
				},
				Boundaries: map[ElementID]Boundary{
					"ent": {ID: "ent", Kind: "Enterprise_Boundary", Label: "Ent"},
					"sb":  {ID: "sb", Kind: "System_Boundary", Label: "SB", Parent: "ent"},
				},
			},
		},
		"skips trivia and unknown macros": {
			Input: "@startuml NAME\n" +
				"!include <C4/C4_Context>\n" +
				"title Hello\n" +
				"caption Cap\n" +
				"' a comment\n" +
				"LAYOUT_WITH_LEGEND()\n" +
				"System(a, \"A\")\n" +
				"note as N\n  body\nend note\n" +
				"UnknownMacro(foo, \"bar\")\n" +
				"SHOW_LEGEND()\n" +
				"@enduml\n",
			Expected: &Diagram{
				Name:       "NAME",
				Order:      []ElementID{"a"},
				Elements:   map[ElementID]Element{"a": {ID: "a", Kind: "System", Label: "A"}},
				Boundaries: map[ElementID]Boundary{},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := Parse(test.Input)
			if err != nil {
				t.Fatalf("Parse returned error: %v", err)
			}
			if !reflect.DeepEqual(test.Expected, got) {
				t.Errorf("Parse mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}

func TestParseMissingStartumlErrors(t *testing.T) {
	if _, err := Parse("Person(u, \"U\")\n@enduml\n"); err == nil {
		t.Error("Parse should error when @startuml is missing")
	}
}

func TestParseExample(t *testing.T) {
	b, err := os.ReadFile("testdata/c4_example.puml")
	if err != nil {
		t.Fatal(err)
	}
	d, err := Parse(string(b))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if got, want := d.Name, "DUMMY-SHOP-C4"; got != want {
		t.Errorf("Name = %q, want %q", got, want)
	}
	if got, want := len(d.Elements), 8; got != want {
		t.Errorf("len(Elements) = %d, want %d", got, want)
	}
	if got, want := len(d.Order), 8; got != want {
		t.Errorf("len(Order) = %d, want %d", got, want)
	}
	if got, want := len(d.Boundaries), 2; got != want {
		t.Errorf("len(Boundaries) = %d, want %d", got, want)
	}
	if got, want := len(d.Rels), 7; got != want {
		t.Errorf("len(Rels) = %d, want %d", got, want)
	}

	wantCourier := Element{ID: "courier", Kind: "Person", External: true, Label: "Courier", Descr: "Delivery provider"}
	if d.Elements["courier"] != wantCourier {
		t.Errorf("Elements[courier] mismatch (-want +got):\n%s", cmp.Diff(wantCourier, d.Elements["courier"]))
	}

	wantOrderSvc := Element{ID: "order_svc", Kind: "System", Label: "Order Service", Descr: "Processes orders", Boundary: "backend"}
	if d.Elements["order_svc"] != wantOrderSvc {
		t.Errorf("Elements[order_svc] mismatch (-want +got):\n%s", cmp.Diff(wantOrderSvc, d.Elements["order_svc"]))
	}

	wantBackend := Boundary{ID: "backend", Kind: "System_Boundary", Label: "Backend Systems", Parent: "shop"}
	if d.Boundaries["backend"] != wantBackend {
		t.Errorf("Boundaries[backend] mismatch (-want +got):\n%s", cmp.Diff(wantBackend, d.Boundaries["backend"]))
	}

	wantFirstRel := Rel{From: "customer", To: "web", Label: "Browses and orders"}
	if d.Rels[0] != wantFirstRel {
		t.Errorf("Rels[0] mismatch (-want +got):\n%s", cmp.Diff(wantFirstRel, d.Rels[0]))
	}
}

func TestSplitArgs(t *testing.T) {
	tests := map[string]struct {
		Input    string
		Expected []string
	}{
		"empty":                {Input: "", Expected: nil},
		"single bare":          {Input: "x", Expected: []string{"x"}},
		"three bare":           {Input: "a, b, c", Expected: []string{"a", "b", "c"}},
		"trims each":           {Input: "  a ,  b ", Expected: []string{"a", "b"}},
		"comma inside quotes":  {Input: `id, "Hello, World"`, Expected: []string{"id", `"Hello, World"`}},
		"escaped quote inside": {Input: `id, "He said \"hi\""`, Expected: []string{"id", `"He said \"hi\""`}},
		"keeps surrounding quotes": {
			Input:    `id, "Order Service (95%, redundant configuration)"`,
			Expected: []string{"id", `"Order Service (95%, redundant configuration)"`},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := splitArgs(test.Input)
			if !slices.Equal(test.Expected, got) {
				t.Errorf("splitArgs mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}

func TestUnquote(t *testing.T) {
	tests := map[string]struct {
		Input    string
		Expected string
	}{
		"bare":              {Input: "id", Expected: "id"},
		"bare trimmed":      {Input: "  id  ", Expected: "id"},
		"quoted":            {Input: `"Hello"`, Expected: "Hello"},
		"quoted with comma": {Input: `"Hello, World"`, Expected: "Hello, World"},
		"escaped quote":     {Input: `"He said \"hi\""`, Expected: `He said "hi"`},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := unquote(test.Input)
			if got != test.Expected {
				t.Errorf("unquote mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}
