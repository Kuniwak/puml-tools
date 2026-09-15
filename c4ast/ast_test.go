package c4ast

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDiagramString(t *testing.T) {
	tests := map[string]struct {
		Diagram  *Diagram
		Expected string
	}{
		"single system element": {
			Diagram: &Diagram{
				Order:    []ElementID{"a"},
				Elements: map[ElementID]Element{"a": {ID: "a", Kind: "System", Label: "A"}},
			},
			Expected: "@startuml\n" +
				"!include <C4/C4_Context>\n\n" +
				"System(a, \"A\")\n" +
				"@enduml\n",
		},
		"named diagram with person": {
			Diagram: &Diagram{
				Name:     "CTX",
				Order:    []ElementID{"u"},
				Elements: map[ElementID]Element{"u": {ID: "u", Kind: "Person", Label: "User"}},
			},
			Expected: "@startuml CTX\n" +
				"!include <C4/C4_Context>\n\n" +
				"Person(u, \"User\")\n" +
				"@enduml\n",
		},
		"external element with description": {
			Diagram: &Diagram{
				Order:    []ElementID{"x"},
				Elements: map[ElementID]Element{"x": {ID: "x", Kind: "System", External: true, Label: "X", Descr: "D"}},
			},
			Expected: "@startuml\n" +
				"!include <C4/C4_Context>\n\n" +
				"System_Ext(x, \"X\", \"D\")\n" +
				"@enduml\n",
		},
		"container element selects the C4_Container include": {
			Diagram: &Diagram{
				Order:    []ElementID{"a"},
				Elements: map[ElementID]Element{"a": {ID: "a", Kind: "Container", Label: "A"}},
			},
			Expected: "@startuml\n" +
				"!include <C4/C4_Container>\n\n" +
				"Container(a, \"A\")\n" +
				"@enduml\n",
		},
		"component element selects the C4_Component include": {
			Diagram: &Diagram{
				Order:    []ElementID{"a"},
				Elements: map[ElementID]Element{"a": {ID: "a", Kind: "ComponentDb", Label: "A"}},
			},
			Expected: "@startuml\n" +
				"!include <C4/C4_Component>\n\n" +
				"ComponentDb(a, \"A\")\n" +
				"@enduml\n",
		},
		"the most specific kind present selects the include": {
			Diagram: &Diagram{
				Order: []ElementID{"s", "c"},
				Elements: map[ElementID]Element{
					"s": {ID: "s", Kind: "System", Label: "S"},
					"c": {ID: "c", Kind: "ContainerQueue", Label: "C"},
				},
			},
			Expected: "@startuml\n" +
				"!include <C4/C4_Container>\n\n" +
				"System(s, \"S\")\n" +
				"ContainerQueue(c, \"C\")\n" +
				"@enduml\n",
		},
		"container boundary selects the C4_Container include": {
			Diagram: &Diagram{
				Order:    []ElementID{"a"},
				Elements: map[ElementID]Element{"a": {ID: "a", Kind: "System", Label: "A", Boundary: "bdy"}},
				Boundaries: map[ElementID]Boundary{
					"bdy": {ID: "bdy", Kind: "Container_Boundary", Label: "Bdy"},
				},
			},
			Expected: "@startuml\n" +
				"!include <C4/C4_Container>\n\n" +
				"Container_Boundary(bdy, \"Bdy\") {\n" +
				"  System(a, \"A\")\n" +
				"}\n" +
				"@enduml\n",
		},
		"elements and rels in declaration order": {
			Diagram: &Diagram{
				Order: []ElementID{"a", "b"},
				Elements: map[ElementID]Element{
					"a": {ID: "a", Kind: "Person", Label: "A"},
					"b": {ID: "b", Kind: "System", Label: "B"},
				},
				Rels: []Rel{{From: "a", To: "b", Label: "uses"}},
			},
			Expected: "@startuml\n" +
				"!include <C4/C4_Context>\n\n" +
				"Person(a, \"A\")\n" +
				"System(b, \"B\")\n\n" +
				"Rel(a, b, \"uses\")\n" +
				"@enduml\n",
		},
		"boundary wraps its members and bare elements stay outside": {
			Diagram: &Diagram{
				Order: []ElementID{"a", "b", "c"},
				Elements: map[ElementID]Element{
					"a": {ID: "a", Kind: "Person", Label: "A"},
					"b": {ID: "b", Kind: "System", Label: "B", Boundary: "bdy"},
					"c": {ID: "c", Kind: "System", Label: "C"},
				},
				Boundaries: map[ElementID]Boundary{
					"bdy": {ID: "bdy", Kind: "System_Boundary", Label: "Bdy"},
				},
			},
			Expected: "@startuml\n" +
				"!include <C4/C4_Context>\n\n" +
				"Person(a, \"A\")\n" +
				"System_Boundary(bdy, \"Bdy\") {\n" +
				"  System(b, \"B\")\n" +
				"}\n" +
				"System(c, \"C\")\n" +
				"@enduml\n",
		},
		"nested boundaries are rendered nested": {
			Diagram: &Diagram{
				Order: []ElementID{"in", "out"},
				Elements: map[ElementID]Element{
					"in":  {ID: "in", Kind: "System", Label: "In", Boundary: "inner"},
					"out": {ID: "out", Kind: "System", Label: "Out", Boundary: "outer"},
				},
				Boundaries: map[ElementID]Boundary{
					"outer": {ID: "outer", Kind: "Enterprise_Boundary", Label: "Outer"},
					"inner": {ID: "inner", Kind: "System_Boundary", Label: "Inner", Parent: "outer"},
				},
			},
			Expected: "@startuml\n" +
				"!include <C4/C4_Context>\n\n" +
				"Enterprise_Boundary(outer, \"Outer\") {\n" +
				"  System_Boundary(inner, \"Inner\") {\n" +
				"    System(in, \"In\")\n" +
				"  }\n" +
				"  System(out, \"Out\")\n" +
				"}\n" +
				"@enduml\n",
		},
		"an element referencing an undeclared boundary surfaces at the top level": {
			Diagram: &Diagram{
				Order:    []ElementID{"a"},
				Elements: map[ElementID]Element{"a": {ID: "a", Kind: "System", Label: "A", Boundary: "ghost"}},
			},
			Expected: "@startuml\n" +
				"!include <C4/C4_Context>\n\n" +
				"System(a, \"A\")\n" +
				"@enduml\n",
		},
		"terminates even when boundary parents form a cycle": {
			Diagram: &Diagram{
				Order:    []ElementID{"a"},
				Elements: map[ElementID]Element{"a": {ID: "a", Kind: "System", Label: "A", Boundary: "x"}},
				Boundaries: map[ElementID]Boundary{
					"x": {ID: "x", Kind: "System_Boundary", Label: "X", Parent: "y"},
					"y": {ID: "y", Kind: "System_Boundary", Label: "Y", Parent: "x"},
				},
			},
			Expected: "@startuml\n" +
				"!include <C4/C4_Context>\n\n" +
				"System(a, \"A\")\n" +
				"@enduml\n",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := test.Diagram.String()
			if got != test.Expected {
				t.Errorf("Diagram.String() mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}
