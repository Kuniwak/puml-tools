package ints_test

import (
	"slices"
	"testing"

	"github.com/Kuniwak/puml-tools/ints"
)

func TestReplaced(t *testing.T) {
	tests := map[string]struct {
		Input    []int
		Index    int
		Value    int
		Expected []int
	}{
		"replaces the first element":              {Input: []int{1, 2, 3}, Index: 0, Value: 9, Expected: []int{9, 2, 3}},
		"replaces a middle element":               {Input: []int{1, 2, 3}, Index: 1, Value: 9, Expected: []int{1, 9, 3}},
		"replaces the last element":               {Input: []int{1, 2, 3}, Index: 2, Value: 9, Expected: []int{1, 2, 9}},
		"returns a copy even with the same value": {Input: []int{1, 2, 3}, Index: 1, Value: 2, Expected: []int{1, 2, 3}},
		"single element":                          {Input: []int{1}, Index: 0, Value: 9, Expected: []int{9}},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			original := slices.Clone(test.Input)

			got := ints.Replaced(test.Input, test.Index, test.Value)

			if !slices.Equal(test.Expected, got) {
				t.Errorf("Replaced(%v, %d, %d) = %v, want %v", original, test.Index, test.Value, got, test.Expected)
			}
			if !slices.Equal(original, test.Input) {
				t.Errorf("Replaced modified its argument: %v, want %v", test.Input, original)
			}
		})
	}
}

func TestKey(t *testing.T) {
	tests := map[string]struct {
		A       []int
		B       []int
		SameKey bool
	}{
		"identical sequences get the same key":                        {A: []int{1, 2, 3}, B: []int{1, 2, 3}, SameKey: true},
		"empty sequences get the same key":                            {A: nil, B: []int{}, SameKey: true},
		"different order gets a different key":                        {A: []int{1, 2}, B: []int{2, 1}, SameKey: false},
		"different length gets a different key":                       {A: []int{1, 2}, B: []int{1, 2, 3}, SameKey: false},
		"same concatenation but different split gets a different key": {A: []int{1, 23}, B: []int{12, 3}, SameKey: false},
		"a negative number is not confused with the separator":        {A: []int{1, -2}, B: []int{1, 2}, SameKey: false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			gotA, gotB := ints.Key(test.A), ints.Key(test.B)
			if (gotA == gotB) != test.SameKey {
				t.Errorf("Key(%v) = %q, Key(%v) = %q: same = %v, want %v", test.A, gotA, test.B, gotB, gotA == gotB, test.SameKey)
			}
		})
	}
}
