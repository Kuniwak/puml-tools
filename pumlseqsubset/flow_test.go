package pumlseqsubset

import (
	"reflect"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestEnabledAlongTrace(t *testing.T) {
	tests := map[string]struct {
		Seq      Seq
		Trace    []int
		Expected [][]int
	}{
		"a sequence can only run its first item": {
			Seq:      NewSeq(Pair{"a", "b"}, Pair{"b", "c"}),
			Trace:    []int{0, 1},
			Expected: [][]int{{0}, {1}, {}},
		},
		"either par branch can run first": {
			Seq:      Seq{Items: []Item{NewPar(NewSeq(Pair{"a", "b"}), NewSeq(Pair{"a", "c"}))}},
			Trace:    []int{1, 0},
			Expected: [][]int{{0, 1}, {0}, {}},
		},
		"a par branch runs sequentially inside itself": {
			Seq: Seq{Items: []Item{NewPar(
				NewSeq(Pair{"a", "b"}, Pair{"b", "c"}),
				NewSeq(Pair{"a", "d"}),
			)}},
			Trace:    []int{0, 2, 1},
			Expected: [][]int{{0, 2}, {1, 2}, {1}, {}},
		},
		"an empty branch is skipped": {
			Seq:      Seq{Items: []Item{NewPar(NewSeq(), NewSeq(Pair{"a", "b"}))}},
			Trace:    []int{0},
			Expected: [][]int{{0}, {}},
		},
		"a par with all-empty branches advances to the next item": {
			Seq: Seq{Items: []Item{
				NewPar(NewSeq(), NewSeq()),
				{Pair: Pair{"a", "b"}},
			}},
			Trace:    []int{0},
			Expected: [][]int{{0}, {}},
		},
		"a nested par is the union of all its branches": {
			Seq: Seq{Items: []Item{NewPar(
				NewSeq(Pair{"x", "y"}),
				Seq{Items: []Item{NewPar(NewSeq(Pair{"a", "b"}), NewSeq(Pair{"c", "d"}))}},
			)}},
			Trace:    []int{1, 2, 0},
			Expected: [][]int{{0, 1, 2}, {0, 2}, {0}, {}},
		},
		"leaving a par advances to the following sequential part": {
			Seq: Seq{Items: []Item{
				NewPar(NewSeq(Pair{"a", "b"}), NewSeq(Pair{"a", "c"})),
				{Pair: Pair{"b", "s"}},
			}},
			Trace:    []int{0, 1, 2},
			Expected: [][]int{{0, 1}, {1}, {2}, {}},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			f := newFlow(test.Seq)
			p := enter(f, 0)

			for step := 0; ; step++ {
				var got []int
				enabled(f, p, &got)
				if !slices.Equal(test.Expected[step], got) {
					t.Fatalf("enabled at step %d mismatch (-want +got):\n%s", step, cmp.Diff(test.Expected[step], got))
				}
				if step == len(test.Trace) {
					break
				}

				next, ok := advance(f, p, test.Trace[step])
				if !ok {
					t.Fatalf("advance(%d) at step %d returned ok=false", test.Trace[step], step)
				}
				p = next
			}

			if !finished(f, p) {
				t.Errorf("finished after the whole trace = false, want true")
			}
		})
	}
}

func TestAdvanceRejectsDisabledEvent(t *testing.T) {
	tests := map[string]struct {
		Seq   Seq
		Index int
	}{
		"the next item in a sequence cannot run yet": {
			Seq:   NewSeq(Pair{"a", "b"}, Pair{"b", "c"}),
			Index: 1,
		},
		"the second item of a par branch cannot run yet": {
			Seq: Seq{Items: []Item{NewPar(
				NewSeq(Pair{"a", "b"}, Pair{"b", "c"}),
				NewSeq(Pair{"a", "d"}),
			)}},
			Index: 1,
		},
		"the sequential part after a par cannot run yet": {
			Seq: Seq{Items: []Item{
				NewPar(NewSeq(Pair{"a", "b"}), NewSeq(Pair{"a", "c"})),
				{Pair: Pair{"b", "s"}},
			}},
			Index: 2,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			f := newFlow(test.Seq)
			p := enter(f, 0)

			got, ok := advance(f, p, test.Index)
			if ok {
				t.Errorf("advance(%d) returned ok=true, want false", test.Index)
			}
			if !reflect.DeepEqual(p, got) {
				t.Errorf("advance(%d) moved the position (-want +got):\n%s", test.Index, cmp.Diff(p, got))
			}
		})
	}
}
