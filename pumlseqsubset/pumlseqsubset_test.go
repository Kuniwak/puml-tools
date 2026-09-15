package pumlseqsubset

import (
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestResidualsAllowSubseq(t *testing.T) {
	tests := map[string]struct {
		Sub      Seq
		Super    Seq
		Expected []int
	}{
		"Nil: sub is empty": {
			Sub:      NewSeq(),
			Super:    NewSeq(Pair{"a", "b"}),
			Expected: nil,
		},
		"both empty": {
			Sub:      NewSeq(),
			Super:    NewSeq(),
			Expected: nil,
		},
		"Keep: exact match": {
			Sub:      NewSeq(Pair{"a", "b"}, Pair{"b", "c"}),
			Super:    NewSeq(Pair{"a", "b"}, Pair{"b", "c"}),
			Expected: nil,
		},
		"Insert: an insertion at the start of super": {
			Sub:      NewSeq(Pair{"b", "c"}),
			Super:    NewSeq(Pair{"a", "b"}, Pair{"b", "c"}),
			Expected: nil,
		},
		"Insert: an insertion in the middle of super": {
			Sub:      NewSeq(Pair{"a", "b"}, Pair{"c", "d"}),
			Super:    NewSeq(Pair{"a", "b"}, Pair{"x", "y"}, Pair{"c", "d"}),
			Expected: nil,
		},
		"one missed item at the end": {
			Sub:      NewSeq(Pair{"a", "b"}, Pair{"z", "w"}),
			Super:    NewSeq(Pair{"a", "b"}),
			Expected: []int{1},
		},
		"one missed item at the start": {
			Sub:      NewSeq(Pair{"z", "w"}, Pair{"a", "b"}),
			Super:    NewSeq(Pair{"a", "b"}),
			Expected: []int{0},
		},
		"minimal set (with repetition)": {
			Sub:      NewSeq(Pair{"X", "Y"}, Pair{"Y", "Z"}, Pair{"Z", "W"}, Pair{"Y", "Z"}),
			Super:    NewSeq(Pair{"X", "Y"}, Pair{"Z", "W"}, Pair{"Y", "Z"}),
			Expected: []int{1},
		},
		"super empty, sub non-empty": {
			Sub:      NewSeq(Pair{"a", "b"}, Pair{"c", "d"}),
			Super:    NewSeq(),
			Expected: []int{0, 1},
		},
		"one leftover after reordering": {
			Sub:      NewSeq(Pair{"a", "b"}, Pair{"c", "d"}, Pair{"a", "b"}),
			Super:    NewSeq(Pair{"c", "d"}, Pair{"a", "b"}),
			Expected: []int{0},
		},
		"a self-message is missing from super": {
			Sub:      NewSeq(Pair{"a", "a"}),
			Super:    NewSeq(Pair{"a", "b"}),
			Expected: []int{0},
		},
		"super is short a duplicate": {
			Sub:      NewSeq(Pair{"a", "b"}, Pair{"a", "b"}),
			Super:    NewSeq(Pair{"a", "b"}),
			Expected: []int{1},
		},
		"a sub listed in the reverse order of super's par branches is still a subsequence": {
			Sub:      NewSeq(Pair{"A", "C"}, Pair{"A", "B"}),
			Super:    Seq{Items: []Item{NewPar(NewSeq(Pair{"A", "B"}), NewSeq(Pair{"A", "C"}))}},
			Expected: nil,
		},
		"any interleaving across branches is a subsequence": {
			Sub: NewSeq(Pair{"A", "D"}, Pair{"A", "B"}, Pair{"D", "E"}, Pair{"B", "C"}),
			Super: Seq{Items: []Item{NewPar(
				NewSeq(Pair{"A", "B"}, Pair{"B", "C"}),
				NewSeq(Pair{"A", "D"}, Pair{"D", "E"}),
			)}},
			Expected: nil,
		},
		"order within the same branch must be preserved": {
			Sub: NewSeq(Pair{"B", "C"}, Pair{"A", "B"}),
			Super: Seq{Items: []Item{NewPar(
				NewSeq(Pair{"A", "B"}, Pair{"B", "C"}),
				NewSeq(Pair{"A", "D"}),
			)}},
			Expected: []int{0},
		},
		"the sequential parts around a par must preserve order": {
			Sub: NewSeq(Pair{"A", "C"}, Pair{"S", "A"}),
			Super: Seq{Items: []Item{
				{Pair: Pair{"S", "A"}},
				NewPar(NewSeq(Pair{"A", "B"}), NewSeq(Pair{"A", "C"})),
			}},
			Expected: []int{0},
		},
		"the par on the sub side is also order-free": {
			Sub:      Seq{Items: []Item{NewPar(NewSeq(Pair{"A", "B"}), NewSeq(Pair{"A", "C"}))}},
			Super:    NewSeq(Pair{"A", "C"}, Pair{"A", "B"}),
			Expected: nil,
		},
		"a nested par is also order-free": {
			Sub: NewSeq(Pair{"C", "D"}, Pair{"X", "Y"}, Pair{"A", "B"}),
			Super: Seq{Items: []Item{NewPar(
				NewSeq(Pair{"X", "Y"}),
				Seq{Items: []Item{NewPar(NewSeq(Pair{"A", "B"}), NewSeq(Pair{"C", "D"}))}},
			)}},
			Expected: nil,
		},
		"missed items in a par are minimal in count": {
			Sub:      NewSeq(Pair{"A", "C"}, Pair{"A", "Z"}, Pair{"A", "B"}),
			Super:    Seq{Items: []Item{NewPar(NewSeq(Pair{"A", "B"}), NewSeq(Pair{"A", "C"}))}},
			Expected: []int{1},
		},
		"a par with an empty branch": {
			Sub:      NewSeq(Pair{"A", "B"}),
			Super:    Seq{Items: []Item{NewPar(NewSeq(), NewSeq(Pair{"A", "B"}))}},
			Expected: nil,
		},
		"removing a hop from a par branch produces a residual": {
			Sub:      NewSeq(Pair{"A", "C"}, Pair{"A", "B"}),
			Super:    Seq{Items: []Item{NewPar(NewSeq(Pair{"A", "B"}), NewSeq())}},
			Expected: []int{0},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			expected := residualsOf(SideSub, test.Sub, test.Expected)

			got := Residuals(ModeAllowSubseq, test.Sub, test.Super)

			if !slices.Equal(expected, got) {
				t.Errorf("Residuals mismatch (-want +got):\n%s", cmp.Diff(expected, got))
			}
		})
	}
}

func TestResidualsExact(t *testing.T) {
	tests := map[string]struct {
		Sub               Seq
		Super             Seq
		ExpectedSubOnly   []int
		ExpectedSuperOnly []int
	}{
		"both empty": {
			Sub:   NewSeq(),
			Super: NewSeq(),
		},
		"exact match": {
			Sub:   NewSeq(Pair{"a", "b"}, Pair{"b", "c"}),
			Super: NewSeq(Pair{"a", "b"}, Pair{"b", "c"}),
		},
		"an insertion at the start of super leaves a super-side residual": {
			Sub:               NewSeq(Pair{"b", "c"}),
			Super:             NewSeq(Pair{"a", "b"}, Pair{"b", "c"}),
			ExpectedSuperOnly: []int{0},
		},
		"an insertion in the middle of super leaves a super-side residual": {
			Sub:               NewSeq(Pair{"a", "b"}, Pair{"c", "d"}),
			Super:             NewSeq(Pair{"a", "b"}, Pair{"x", "y"}, Pair{"c", "d"}),
			ExpectedSuperOnly: []int{1},
		},
		"a message only in sub is a sub-side residual": {
			Sub:             NewSeq(Pair{"a", "b"}, Pair{"z", "w"}),
			Super:           NewSeq(Pair{"a", "b"}),
			ExpectedSubOnly: []int{1},
		},
		"both sides have leftovers": {
			Sub:               NewSeq(Pair{"a", "b"}, Pair{"b", "x"}, Pair{"b", "a"}),
			Super:             NewSeq(Pair{"a", "b"}, Pair{"b", "y"}, Pair{"b", "a"}),
			ExpectedSubOnly:   []int{1},
			ExpectedSuperOnly: []int{1},
		},
		"sub empty, super non-empty": {
			Sub:               NewSeq(),
			Super:             NewSeq(Pair{"a", "b"}, Pair{"c", "d"}),
			ExpectedSuperOnly: []int{0, 1},
		},
		"super empty, sub non-empty": {
			Sub:             NewSeq(Pair{"a", "b"}, Pair{"c", "d"}),
			Super:           NewSeq(),
			ExpectedSubOnly: []int{0, 1},
		},
		"the same set in a different order still produces residuals": {
			Sub:               NewSeq(Pair{"a", "b"}, Pair{"c", "d"}),
			Super:             NewSeq(Pair{"c", "d"}, Pair{"a", "b"}),
			ExpectedSubOnly:   []int{0},
			ExpectedSuperOnly: []int{1},
		},
		"super is short a duplicate": {
			Sub:             NewSeq(Pair{"a", "b"}, Pair{"a", "b"}),
			Super:           NewSeq(Pair{"a", "b"}),
			ExpectedSubOnly: []int{1},
		},
		"sub is short a duplicate": {
			Sub:               NewSeq(Pair{"a", "b"}),
			Super:             NewSeq(Pair{"a", "b"}, Pair{"a", "b"}),
			ExpectedSuperOnly: []int{1},
		},
		"a sub listed in the reverse order of par branches counts as a match": {
			Sub:   NewSeq(Pair{"A", "C"}, Pair{"A", "B"}),
			Super: Seq{Items: []Item{NewPar(NewSeq(Pair{"A", "B"}), NewSeq(Pair{"A", "C"}))}},
		},
		"the par on the sub side is also order-free": {
			Sub:   Seq{Items: []Item{NewPar(NewSeq(Pair{"A", "B"}), NewSeq(Pair{"A", "C"}))}},
			Super: NewSeq(Pair{"A", "C"}, Pair{"A", "B"}),
		},
		"order within the same branch must be preserved": {
			Sub: NewSeq(Pair{"B", "C"}, Pair{"A", "B"}),
			Super: Seq{Items: []Item{NewPar(
				NewSeq(Pair{"A", "B"}, Pair{"B", "C"}),
				NewSeq(Pair{"A", "D"}),
			)}},
			ExpectedSubOnly:   []int{0},
			ExpectedSuperOnly: []int{1, 2},
		},
		"a message missing from the par branches is a sub-side residual": {
			Sub:             NewSeq(Pair{"A", "C"}, Pair{"A", "Z"}, Pair{"A", "B"}),
			Super:           Seq{Items: []Item{NewPar(NewSeq(Pair{"A", "B"}), NewSeq(Pair{"A", "C"}))}},
			ExpectedSubOnly: []int{1},
		},
		"a message only in the par branches is a super-side residual": {
			Sub:               NewSeq(Pair{"A", "B"}),
			Super:             Seq{Items: []Item{NewPar(NewSeq(Pair{"A", "B"}), NewSeq(Pair{"A", "C"}))}},
			ExpectedSuperOnly: []int{1},
		},
		"swapping sub and super swaps the residual too (leftover on sub side)": {
			Sub:             NewSeq(Pair{"a", "b"}, Pair{"c", "d"}, Pair{"e", "f"}),
			Super:           NewSeq(Pair{"a", "b"}, Pair{"e", "f"}),
			ExpectedSubOnly: []int{1},
		},
		"swapping sub and super swaps the residual too (leftover on super side)": {
			Sub:               NewSeq(Pair{"a", "b"}, Pair{"e", "f"}),
			Super:             NewSeq(Pair{"a", "b"}, Pair{"c", "d"}, Pair{"e", "f"}),
			ExpectedSuperOnly: []int{1},
		},
		"on a tie, the later item in document order becomes the residual": {
			Sub:             NewSeq(Pair{"X", "Y"}, Pair{"Y", "Z"}, Pair{"Z", "W"}, Pair{"Y", "Z"}),
			Super:           NewSeq(Pair{"X", "Y"}, Pair{"Z", "W"}, Pair{"Y", "Z"}),
			ExpectedSubOnly: []int{1},
		},
		"when both sides are par, either linearization is allowed": {
			Sub:   Seq{Items: []Item{NewPar(NewSeq(Pair{"A", "B"}), NewSeq(Pair{"A", "C"}))}},
			Super: Seq{Items: []Item{NewPar(NewSeq(Pair{"A", "C"}), NewSeq(Pair{"A", "B"}))}},
		},
		"even when both sides are par, leftovers become residuals": {
			Sub: Seq{Items: []Item{NewPar(
				NewSeq(Pair{"A", "B"}, Pair{"A", "Z"}),
				NewSeq(Pair{"A", "C"}),
			)}},
			Super:           Seq{Items: []Item{NewPar(NewSeq(Pair{"A", "C"}), NewSeq(Pair{"A", "B"}))}},
			ExpectedSubOnly: []int{1},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			expected := append(
				residualsOf(SideSub, test.Sub, test.ExpectedSubOnly),
				residualsOf(SideSuper, test.Super, test.ExpectedSuperOnly)...,
			)

			got := Residuals(ModeExact, test.Sub, test.Super)

			if !slices.Equal(expected, got) {
				t.Errorf("Residuals mismatch (-want +got):\n%s", cmp.Diff(expected, got))
			}
		})
	}
}

func TestPairs(t *testing.T) {
	tests := map[string]struct {
		Seq      Seq
		Expected []Pair
	}{
		"empty": {
			Seq:      NewSeq(),
			Expected: nil,
		},
		"NewSeq preserves argument order": {
			Seq:      NewSeq(Pair{"a", "b"}, Pair{"b", "c"}),
			Expected: []Pair{{"a", "b"}, {"b", "c"}},
		},
		"par branches are ordered as written": {
			Seq: Seq{Items: []Item{
				{Pair: Pair{"S", "A"}},
				NewPar(NewSeq(Pair{"A", "B"}), NewSeq(Pair{"A", "C"}, Pair{"C", "D"})),
			}},
			Expected: []Pair{{"S", "A"}, {"A", "B"}, {"A", "C"}, {"C", "D"}},
		},
		"a nested par is also ordered as written": {
			Seq: Seq{Items: []Item{NewPar(
				NewSeq(Pair{"X", "Y"}),
				Seq{Items: []Item{NewPar(NewSeq(Pair{"A", "B"}), NewSeq(Pair{"C", "D"}))}},
			)}},
			Expected: []Pair{{"X", "Y"}, {"A", "B"}, {"C", "D"}},
		},
		"a par with no branches": {
			Seq:      Seq{Items: []Item{NewPar()}},
			Expected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := Pairs(test.Seq)
			if !slices.Equal(test.Expected, got) {
				t.Errorf("Pairs mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}

func residualsOf(side Side, s Seq, idxs []int) []Residual {
	pairs := Pairs(s)
	out := make([]Residual, 0, len(idxs))
	for _, i := range idxs {
		out = append(out, Residual{Side: side, Index: i, Pair: pairs[i]})
	}
	return out
}

func TestParseMode(t *testing.T) {
	tests := map[string]struct {
		Input       string
		Expected    Mode
		ExpectedErr bool
	}{
		"exact":                       {Input: "exact", Expected: ModeExact},
		"allow-subseq":                {Input: "allow-subseq", Expected: ModeAllowSubseq},
		"an unknown name is an error": {Input: "subseq", ExpectedErr: true},
		"an empty string is an error": {Input: "", ExpectedErr: true},
		"uppercase is treated as a different, invalid name": {Input: "EXACT", ExpectedErr: true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ParseMode(test.Input)
			if test.ExpectedErr {
				if err == nil {
					t.Fatalf("ParseMode(%q) should fail, but got %q", test.Input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseMode(%q) should succeed, but got error: %s", test.Input, err)
			}
			if got != test.Expected {
				t.Errorf("ParseMode(%q) = %q, want %q", test.Input, got, test.Expected)
			}
		})
	}
}

func TestRecord(t *testing.T) {
	tests := map[string]struct {
		Mode     Mode
		Residual Residual
		Expected []string
	}{
		"exact's sub side has 3 columns": {
			Mode:     ModeExact,
			Residual: Residual{Side: SideSub, Index: 0, Pair: Pair{"a", "b"}},
			Expected: []string{"sub", "a", "b"},
		},
		"exact's super side has 3 columns": {
			Mode:     ModeExact,
			Residual: Residual{Side: SideSuper, Index: 1, Pair: Pair{"b", "c"}},
			Expected: []string{"super", "b", "c"},
		},
		"allow-subseq has 2 columns": {
			Mode:     ModeAllowSubseq,
			Residual: Residual{Side: SideSub, Index: 0, Pair: Pair{"a", "b"}},
			Expected: []string{"a", "b"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := Record(test.Mode, test.Residual)
			if diff := cmp.Diff(test.Expected, got); diff != "" {
				t.Errorf("Record(%q, %+v) mismatch (-want +got):\n%s", test.Mode, test.Residual, diff)
			}
		})
	}
}

func TestRecordPanicsOnUnknownMode(t *testing.T) {
	tests := map[string]struct {
		Mode Mode
	}{
		"zero value":    {Mode: ""},
		"unknown value": {Mode: "unknown"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("Record(%q, ...) should panic, but returned", test.Mode)
				}
			}()
			Record(test.Mode, Residual{Side: SideSub, Index: 0, Pair: Pair{"a", "b"}})
		})
	}
}
