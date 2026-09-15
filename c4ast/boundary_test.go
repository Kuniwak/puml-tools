package c4ast

import (
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestBoundaryMembers(t *testing.T) {
	nested := "@startuml\n" +
		"Enterprise_Boundary(ent, \"Ent\") {\n" +
		"  System(s1, \"S1\")\n" +
		"  System_Boundary(sb, \"SB\") {\n" +
		"    System(s2, \"S2\")\n" +
		"  }\n" +
		"}\n" +
		"System(top, \"Top\")\n" +
		"@enduml\n"

	tests := map[string]struct {
		Input    string
		Boundary ElementID
		Want     []ElementID
		WantErr  bool
	}{
		"single boundary two members": {
			Input: "@startuml\n" +
				"System_Boundary(sb, \"SB\") {\n" +
				"  System(a, \"A\")\n" +
				"  System(b, \"B\")\n" +
				"}\n" +
				"@enduml\n",
			Boundary: "sb",
			Want:     []ElementID{"a", "b"},
		},
		"nested transitive from outer": {
			Input:    nested,
			Boundary: "ent",
			Want:     []ElementID{"s1", "s2"},
		},
		"nested inner only": {
			Input:    nested,
			Boundary: "sb",
			Want:     []ElementID{"s2"},
		},
		"empty boundary": {
			Input: "@startuml\n" +
				"System_Boundary(sb, \"SB\") {\n" +
				"}\n" +
				"System(top, \"Top\")\n" +
				"@enduml\n",
			Boundary: "sb",
			Want:     nil,
		},
		"boundary not found": {
			Input:    nested,
			Boundary: "nope",
			WantErr:  true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			d, err := Parse(test.Input)
			if err != nil {
				t.Fatalf("Parse returned error: %v", err)
			}
			got, err := BoundaryMembers(d, test.Boundary)
			if test.WantErr {
				if err == nil {
					t.Fatalf("BoundaryMembers should error for boundary %q", test.Boundary)
				}
				return
			}
			if err != nil {
				t.Fatalf("BoundaryMembers returned error: %v", err)
			}
			if !slices.Equal(test.Want, got) {
				t.Errorf("BoundaryMembers mismatch (-want +got):\n%s", cmp.Diff(test.Want, got))
			}
		})
	}
}

func TestEnclosingBoundary(t *testing.T) {
	d := &Diagram{
		Order: []ElementID{"in", "ghost", "looped"},
		Elements: map[ElementID]Element{
			"in":     {ID: "in", Kind: "System", Boundary: "inner"},
			"ghost":  {ID: "ghost", Kind: "System", Boundary: "nowhere"},
			"looped": {ID: "looped", Kind: "System", Boundary: "x"},
		},
		Boundaries: map[ElementID]Boundary{
			"outer": {ID: "outer", Kind: "System_Boundary"},
			"inner": {ID: "inner", Kind: "System_Boundary", Parent: "outer"},
			"x":     {ID: "x", Kind: "System_Boundary", Parent: "y"},
			"y":     {ID: "y", Kind: "System_Boundary", Parent: "x"},
		},
	}

	tests := map[string]struct {
		ID       ElementID
		Expected ElementID
	}{
		"a nested element returns its immediately enclosing boundary": {ID: "in", Expected: "inner"},
		"a boundary returns its parent":                               {ID: "inner", Expected: "outer"},
		"a top-level boundary returns empty":                          {ID: "outer", Expected: ""},
		"an element referencing an undeclared boundary returns empty": {ID: "ghost", Expected: ""},
		"an element referencing a cyclic boundary returns empty":      {ID: "looped", Expected: ""},
		"a cyclic boundary itself also returns empty":                 {ID: "x", Expected: ""},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := d.EnclosingBoundary(test.ID); got != test.Expected {
				t.Errorf("EnclosingBoundary(%q) = %q, want %q", test.ID, got, test.Expected)
			}
		})
	}
}
