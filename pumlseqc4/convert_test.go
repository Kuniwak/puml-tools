package pumlseqc4

import (
	"reflect"
	"testing"

	"github.com/Kuniwak/puml-tools/c4ast"
	"github.com/Kuniwak/puml-tools/pumlseqast"
	"github.com/google/go-cmp/cmp"
)

func TestFromSeq(t *testing.T) {
	tests := map[string]struct {
		Input    string
		Expected *c4ast.Diagram
	}{
		"participants and message": {
			Input: "@startuml\nactor a\nparticipant \"Bee\" as b\na -> b : hi\n@enduml\n",
			Expected: &c4ast.Diagram{
				Order: []c4ast.ElementID{"a", "b"},
				Elements: map[c4ast.ElementID]c4ast.Element{
					"a": {ID: "a", Kind: "Person", Label: "a"},
					"b": {ID: "b", Kind: "System", Label: "Bee"},
				},
				Boundaries: map[c4ast.ElementID]c4ast.Boundary{},
				Rels:       []c4ast.Rel{{From: "a", To: "b", Label: "hi"}},
			},
		},
		"self message dropped": {
			Input: "@startuml\na -> a : x\na -> b : y\n@enduml\n",
			Expected: &c4ast.Diagram{
				Order: []c4ast.ElementID{"a", "b"},
				Elements: map[c4ast.ElementID]c4ast.Element{
					"a": {ID: "a", Kind: "System", Label: "a"},
					"b": {ID: "b", Kind: "System", Label: "b"},
				},
				Boundaries: map[c4ast.ElementID]c4ast.Boundary{},
				Rels:       []c4ast.Rel{{From: "a", To: "b", Label: "y"}},
			},
		},
		"duplicate triple collapsed": {
			Input: "@startuml\na -> b : x\na -> b : x\na -> b : z\n@enduml\n",
			Expected: &c4ast.Diagram{
				Order: []c4ast.ElementID{"a", "b"},
				Elements: map[c4ast.ElementID]c4ast.Element{
					"a": {ID: "a", Kind: "System", Label: "a"},
					"b": {ID: "b", Kind: "System", Label: "b"},
				},
				Boundaries: map[c4ast.ElementID]c4ast.Boundary{},
				Rels: []c4ast.Rel{
					{From: "a", To: "b", Label: "x"},
					{From: "a", To: "b", Label: "z"},
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			d, err := pumlseqast.Parse(test.Input)
			if err != nil {
				t.Fatal(err)
			}
			got := FromSeq(d)
			if !reflect.DeepEqual(test.Expected, got) {
				t.Errorf("FromSeq mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}

func TestFromSeqs(t *testing.T) {
	tests := map[string]struct {
		Inputs   []string
		Expected *c4ast.Diagram
	}{
		"empty input": {
			Inputs: nil,
			Expected: &c4ast.Diagram{
				Elements:   map[c4ast.ElementID]c4ast.Element{},
				Boundaries: map[c4ast.ElementID]c4ast.Boundary{},
			},
		},
		"union dedups elements and rels across diagrams": {
			Inputs: []string{
				"@startuml\na -> b : x\n@enduml\n",
				"@startuml\nb -> c : y\na -> b : x\n@enduml\n",
			},
			Expected: &c4ast.Diagram{
				Order: []c4ast.ElementID{"a", "b", "c"},
				Elements: map[c4ast.ElementID]c4ast.Element{
					"a": {ID: "a", Kind: "System", Label: "a"},
					"b": {ID: "b", Kind: "System", Label: "b"},
					"c": {ID: "c", Kind: "System", Label: "c"},
				},
				Boundaries: map[c4ast.ElementID]c4ast.Boundary{},
				Rels: []c4ast.Rel{
					{From: "a", To: "b", Label: "x"},
					{From: "b", To: "c", Label: "y"},
				},
			},
		},
		"element kind first-wins across diagrams": {
			Inputs: []string{
				"@startuml\nactor a\na -> b : x\n@enduml\n",
				"@startuml\nparticipant a\n@enduml\n",
			},
			Expected: &c4ast.Diagram{
				Order: []c4ast.ElementID{"a", "b"},
				Elements: map[c4ast.ElementID]c4ast.Element{
					"a": {ID: "a", Kind: "Person", Label: "a"},
					"b": {ID: "b", Kind: "System", Label: "b"},
				},
				Boundaries: map[c4ast.ElementID]c4ast.Boundary{},
				Rels:       []c4ast.Rel{{From: "a", To: "b", Label: "x"}},
			},
		},
		"carries no diagram name even when inputs are named": {
			Inputs: []string{"@startuml SEQ-01\na -> b : x\n@enduml\n"},
			Expected: &c4ast.Diagram{
				Order: []c4ast.ElementID{"a", "b"},
				Elements: map[c4ast.ElementID]c4ast.Element{
					"a": {ID: "a", Kind: "System", Label: "a"},
					"b": {ID: "b", Kind: "System", Label: "b"},
				},
				Boundaries: map[c4ast.ElementID]c4ast.Boundary{},
				Rels:       []c4ast.Rel{{From: "a", To: "b", Label: "x"}},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var ds []*pumlseqast.Diagram
			for _, in := range test.Inputs {
				d, err := pumlseqast.Parse(in)
				if err != nil {
					t.Fatal(err)
				}
				ds = append(ds, d)
			}
			got := FromSeqs(ds)
			if got.Name != "" {
				t.Errorf("Name = %q, want empty", got.Name)
			}
			if !reflect.DeepEqual(test.Expected, got) {
				t.Errorf("FromSeqs mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}

func TestFromSeqsWithBoxBoundaries(t *testing.T) {
	tests := map[string]struct {
		Inputs   []string
		Expected *c4ast.Diagram
		WantErr  bool
	}{
		"box becomes a boundary and participants outside it stay outside": {
			Inputs: []string{"@startuml\nactor user\nbox \"Order Service\"\n  participant tms_back\nend box\nuser -> tms_back : ask\n@enduml\n"},
			Expected: &c4ast.Diagram{
				Order: []c4ast.ElementID{"user", "tms_back"},
				Elements: map[c4ast.ElementID]c4ast.Element{
					"user":     {ID: "user", Kind: "Person", Label: "user"},
					"tms_back": {ID: "tms_back", Kind: "System", Label: "tms_back", Boundary: "order_service"},
				},
				Boundaries: map[c4ast.ElementID]c4ast.Boundary{
					"order_service": {ID: "order_service", Kind: "System_Boundary", Label: "Order Service"},
				},
				Rels: []c4ast.Rel{{From: "user", To: "tms_back", Label: "ask"}},
			},
		},
		"decoration tags on a box label are dropped on both ends": {
			Inputs: []string{"@startuml\nbox \"<back:#424242><color:#FFFFFF>  Order Service  </color></back>\"\n  participant a\nend box\n@enduml\n"},
			Expected: &c4ast.Diagram{
				Order: []c4ast.ElementID{"a"},
				Elements: map[c4ast.ElementID]c4ast.Element{
					"a": {ID: "a", Kind: "System", Label: "a", Boundary: "order_service"},
				},
				Boundaries: map[c4ast.ElementID]c4ast.Boundary{
					"order_service": {ID: "order_service", Kind: "System_Boundary", Label: "Order Service"},
				},
			},
		},
		"a box without a display name creates no boundary": {
			Inputs: []string{"@startuml\nbox #AliceBlue\n  participant a\nend box\n@enduml\n"},
			Expected: &c4ast.Diagram{
				Order: []c4ast.ElementID{"a"},
				Elements: map[c4ast.ElementID]c4ast.Element{
					"a": {ID: "a", Kind: "System", Label: "a"},
				},
				Boundaries: map[c4ast.ElementID]c4ast.Boundary{},
			},
		},
		"box membership across multiple inputs is unioned": {
			Inputs: []string{
				"@startuml\nbox \"X\"\n  participant a\nend box\n@enduml\n",
				"@startuml\nbox \"Y\"\n  participant b\nend box\na -> b : m\n@enduml\n",
			},
			Expected: &c4ast.Diagram{
				Order: []c4ast.ElementID{"a", "b"},
				Elements: map[c4ast.ElementID]c4ast.Element{
					"a": {ID: "a", Kind: "System", Label: "a", Boundary: "x"},
					"b": {ID: "b", Kind: "System", Label: "b", Boundary: "y"},
				},
				Boundaries: map[c4ast.ElementID]c4ast.Boundary{
					"x": {ID: "x", Kind: "System_Boundary", Label: "X"},
					"y": {ID: "y", Kind: "System_Boundary", Label: "Y"},
				},
				Rels: []c4ast.Rel{{From: "a", To: "b", Label: "m"}},
			},
		},
		"error when different box labels map to the same boundary id": {
			Inputs:  []string{"@startuml\nbox \"Order Service\"\n  participant a\nend box\nbox \"Order/Service\"\n  participant b\nend box\n@enduml\n"},
			WantErr: true,
		},
		"error when the same id belongs to a different box across diagrams": {
			Inputs: []string{
				"@startuml\nbox \"X\"\n  participant a\nend box\n@enduml\n",
				"@startuml\nbox \"Y\"\n  participant a\nend box\n@enduml\n",
			},
			WantErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ds := make([]*pumlseqast.Diagram, 0, len(test.Inputs))
			for _, in := range test.Inputs {
				d, err := pumlseqast.Parse(in)
				if err != nil {
					t.Fatal(err)
				}
				ds = append(ds, d)
			}
			got, err := FromSeqsWithBoxBoundaries(ds)
			if test.WantErr {
				if err == nil {
					t.Fatalf("want error, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(test.Expected, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}

func TestBoundaryID(t *testing.T) {
	tests := map[string]struct {
		Label    string
		Expected c4ast.ElementID
	}{
		"ASCII label becomes a readable slug":                {Label: "Order Service", Expected: "order_service"},
		"symbols collapse to _ and are trimmed at both ends": {Label: " Order/Service! ", Expected: "order_service"},
		"non-ASCII label becomes a deterministic hash id":    {Label: "Café Ops", Expected: "bdy_e2eac237"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := BoundaryID(test.Label); got != test.Expected {
				t.Errorf("BoundaryID(%q) = %q, want %q", test.Label, got, test.Expected)
			}
		})
	}
}
