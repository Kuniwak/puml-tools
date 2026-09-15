package sets

import "testing"

func TestSet(t *testing.T) {
	tests := map[string]struct {
		Initial  []string
		Add      []string
		Query    string
		Expected bool
	}{
		"contains initial member": {
			Initial:  []string{"a", "b"},
			Query:    "a",
			Expected: true,
		},
		"does not contain missing member": {
			Initial:  []string{"a", "b"},
			Query:    "c",
			Expected: false,
		},
		"contains added member": {
			Initial:  []string{"a"},
			Add:      []string{"c"},
			Query:    "c",
			Expected: true,
		},
		"empty set contains nothing": {
			Query:    "a",
			Expected: false,
		},
		"deduplicates initial members": {
			Initial:  []string{"a", "a"},
			Query:    "a",
			Expected: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			s := New(test.Initial...)
			for _, item := range test.Add {
				s.Add(item)
			}
			if got := s.Contains(test.Query); got != test.Expected {
				t.Errorf("Contains(%q) = %v, want %v", test.Query, got, test.Expected)
			}
		})
	}
}

func TestAddAll(t *testing.T) {
	tests := map[string]struct {
		Initial  []string
		Other    []string
		Expected []string
	}{
		"adds members of the other set": {
			Initial:  []string{"a"},
			Other:    []string{"b", "c"},
			Expected: []string{"a", "b", "c"},
		},
		"keeps members already present": {
			Initial:  []string{"a", "b"},
			Other:    []string{"b"},
			Expected: []string{"a", "b"},
		},
		"empty other keeps the receiver": {
			Initial:  []string{"a"},
			Expected: []string{"a"},
		},
		"empty receiver takes the other": {
			Other:    []string{"a"},
			Expected: []string{"a"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			s := New(test.Initial...)
			s.AddAll(New(test.Other...))
			if len(s) != len(test.Expected) {
				t.Errorf("set size = %d, want %d", len(s), len(test.Expected))
			}
			for _, item := range test.Expected {
				if !s.Contains(item) {
					t.Errorf("Contains(%q) = false, want true", item)
				}
			}
		})
	}
}

func TestSorted(t *testing.T) {
	tests := map[string]struct {
		Initial  []string
		Expected []string
	}{
		"sorts members in ascending order": {
			Initial:  []string{"c", "a", "b"},
			Expected: []string{"a", "b", "c"},
		},
		"deduplicates members": {
			Initial:  []string{"a", "a"},
			Expected: []string{"a"},
		},
		"empty set is an empty slice": {
			Initial:  nil,
			Expected: []string{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := New(test.Initial...).Sorted()
			if len(got) != len(test.Expected) {
				t.Fatalf("Sorted() = %v, want %v", got, test.Expected)
			}
			for i, want := range test.Expected {
				if got[i] != want {
					t.Errorf("Sorted()[%d] = %q, want %q", i, got[i], want)
				}
			}
		})
	}
}

func TestIntersects(t *testing.T) {
	tests := map[string]struct {
		A        []string
		B        []string
		Expected bool
	}{
		"sharing one member": {
			A:        []string{"a", "b"},
			B:        []string{"b", "c"},
			Expected: true,
		},
		"disjoint members": {
			A:        []string{"a", "b"},
			B:        []string{"c", "d"},
			Expected: false,
		},
		"empty on the left": {
			B:        []string{"a"},
			Expected: false,
		},
		"empty on the right": {
			A:        []string{"a"},
			Expected: false,
		},
		"both empty": {
			Expected: false,
		},
		"the smaller set on the right": {
			A:        []string{"a", "b", "c"},
			B:        []string{"c"},
			Expected: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := Intersects(New(test.A...), New(test.B...)); got != test.Expected {
				t.Errorf("Intersects(%v, %v) = %v, want %v", test.A, test.B, got, test.Expected)
			}
		})
	}
}
