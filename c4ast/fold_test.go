package c4ast

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFold(t *testing.T) {
	base := "@startuml\n" +
		"System_Boundary(shop, \"Shop\") {\n" +
		"  System(web, \"Web\")\n" +
		"  System(api, \"API\")\n" +
		"}\n" +
		"System_Ext(pay, \"Pay\")\n" +
		"Person(user, \"User\")\n" +
		"Rel(user, web, \"browse\")\n" +
		"Rel(web, api, \"call\")\n" +
		"Rel(api, pay, \"charge\")\n" +
		"Rel(user, api, \"order\")\n" +
		"@enduml\n"

	tests := map[string]struct {
		Input   string
		IDs     []string
		NewID   string
		NewDesc string
		Want    *Diagram
		WantErr bool
	}{
		"fold remaps rels, drops self, keeps distinct labels": {
			Input:   base,
			IDs:     []string{"web", "api"},
			NewID:   "shop",
			NewDesc: "Shop",
			Want: &Diagram{
				Order: []ElementID{"shop", "pay", "user"},
				Elements: map[ElementID]Element{
					"shop": {ID: "shop", Kind: "System", Label: "Shop"},
					"pay":  {ID: "pay", Kind: "System", External: true, Label: "Pay"},
					"user": {ID: "user", Kind: "Person", Label: "User"},
				},
				Rels: []Rel{
					{From: "user", To: "shop", Label: "browse"},
					{From: "shop", To: "pay", Label: "charge"},
					{From: "user", To: "shop", Label: "order"},
				},
			},
		},
		"elements surviving the fold keep no boundary reference": {
			Input:   base,
			IDs:     []string{"pay"},
			NewID:   "pay2",
			NewDesc: "Pay2",
			Want: &Diagram{
				Order: []ElementID{"web", "api", "pay2", "user"},
				Elements: map[ElementID]Element{
					"web":  {ID: "web", Kind: "System", Label: "Web"},
					"api":  {ID: "api", Kind: "System", Label: "API"},
					"pay2": {ID: "pay2", Kind: "System", Label: "Pay2"},
					"user": {ID: "user", Kind: "Person", Label: "User"},
				},
				Rels: []Rel{
					{From: "user", To: "web", Label: "browse"},
					{From: "web", To: "api", Label: "call"},
					{From: "api", To: "pay2", Label: "charge"},
					{From: "user", To: "api", Label: "order"},
				},
			},
		},
		"empty new-desc falls back to new-id as label": {
			Input: base,
			IDs:   []string{"web", "api"},
			NewID: "shop",
			Want: &Diagram{
				Order: []ElementID{"shop", "pay", "user"},
				Elements: map[ElementID]Element{
					"shop": {ID: "shop", Kind: "System", Label: "shop"},
					"pay":  {ID: "pay", Kind: "System", External: true, Label: "Pay"},
					"user": {ID: "user", Kind: "Person", Label: "User"},
				},
				Rels: []Rel{
					{From: "user", To: "shop", Label: "browse"},
					{From: "shop", To: "pay", Label: "charge"},
					{From: "user", To: "shop", Label: "order"},
				},
			},
		},
		"identical rels dedup after remap": {
			Input: "@startuml\n" +
				"System_Boundary(shop, \"Shop\") {\n" +
				"  System(web, \"Web\")\n" +
				"  System(api, \"API\")\n" +
				"}\n" +
				"Person(user, \"User\")\n" +
				"Rel(user, web, \"x\")\n" +
				"Rel(user, api, \"x\")\n" +
				"@enduml\n",
			IDs:   []string{"web", "api"},
			NewID: "shop",
			Want: &Diagram{
				Order: []ElementID{"shop", "user"},
				Elements: map[ElementID]Element{
					"shop": {ID: "shop", Kind: "System", Label: "shop"},
					"user": {ID: "user", Kind: "Person", Label: "User"},
				},
				Rels: []Rel{
					{From: "user", To: "shop", Label: "x"},
				},
			},
		},
		"empty ids errors": {
			Input:   base,
			IDs:     nil,
			NewID:   "shop",
			WantErr: true,
		},
		"unknown id errors": {
			Input:   base,
			IDs:     []string{"web", "nope"},
			NewID:   "shop",
			WantErr: true,
		},
		"new-id conflicts with non-folded element errors": {
			Input:   base,
			IDs:     []string{"web", "api"},
			NewID:   "user",
			WantErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			d, err := Parse(test.Input)
			if err != nil {
				t.Fatalf("Parse returned error: %v", err)
			}
			got, err := Fold(d, test.IDs, test.NewID, test.NewDesc)
			if test.WantErr {
				if err == nil {
					t.Fatalf("Fold should error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Fold returned error: %v", err)
			}
			if !reflect.DeepEqual(test.Want, got) {
				t.Errorf("Fold mismatch (-want +got):\n%s", cmp.Diff(test.Want, got))
			}
		})
	}
}
