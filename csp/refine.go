package csp

import (
	"slices"
	"sort"

	"github.com/Kuniwak/puml-tools/ints"
	"github.com/Kuniwak/puml-tools/pumlseqast"
)

type Result struct {
	Refines   bool
	Witness   []pumlseqast.MessageKey
	Offending pumlseqast.MessageKey
}

type TraceDFA struct {
	start int
	trans []map[pumlseqast.MessageKey]int
}

func (d *TraceDFA) Start() int { return d.start }

func (d *TraceDFA) NumStates() int { return len(d.trans) }

func (d *TraceDFA) Next(state int, label pumlseqast.MessageKey) (next int, ok bool) {
	next, ok = d.trans[state][label]
	return next, ok
}

func (d *TraceDFA) Labels(state int) []pumlseqast.MessageKey {
	return sortedLabels(d.trans[state])
}

func RefinesTraces(impl, spec *LTS) Result {
	di := impl.DeterminizeTraces()
	ds := spec.DeterminizeTraces()

	type pair struct{ qi, qs int }
	type back struct {
		prev  pair
		label pumlseqast.MessageKey
		root  bool
	}
	start := pair{qi: di.start, qs: ds.start}
	parent := map[pair]back{start: {root: true}}
	queue := []pair{start}

	reconstruct := func(p pair) []pumlseqast.MessageKey {
		var rev []pumlseqast.MessageKey
		for {
			b := parent[p]
			if b.root {
				break
			}
			rev = append(rev, b.label)
			p = b.prev
		}
		for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
			rev[i], rev[j] = rev[j], rev[i]
		}
		return rev
	}

	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		for _, label := range sortedLabels(di.trans[p.qi]) {
			ts, ok := ds.trans[p.qs][label]
			if !ok {
				return Result{Refines: false, Witness: reconstruct(p), Offending: label}
			}
			np := pair{qi: di.trans[p.qi][label], qs: ts}
			if _, seen := parent[np]; !seen {
				parent[np] = back{prev: p, label: label}
				queue = append(queue, np)
			}
		}
	}
	return Result{Refines: true}
}

func (l *LTS) DeterminizeTraces() *TraceDFA {
	d := &TraceDFA{}
	index := map[string]int{}
	var queue [][]int
	getState := func(set []int) int {
		key := ints.Key(set)
		if id, ok := index[key]; ok {
			return id
		}
		id := len(d.trans)
		d.trans = append(d.trans, map[pumlseqast.MessageKey]int{})
		index[key] = id
		queue = append(queue, set)
		return id
	}

	d.start = getState(l.EpsClosure([]int{l.start}))
	for len(queue) > 0 {
		set := queue[0]
		queue = queue[1:]
		id := index[ints.Key(set)]
		byLabel := map[pumlseqast.MessageKey][]int{}
		for _, s := range set {
			for _, mv := range l.moves[s] {
				byLabel[mv.Label] = append(byLabel[mv.Label], mv.To)
			}
		}
		for label, tos := range byLabel {
			d.trans[id][label] = getState(l.EpsClosure(tos))
		}
	}
	return d
}

func (l *LTS) EpsClosure(states []int) []int {
	for _, s := range states {
		l.mustState(s)
	}
	seen := map[int]bool{}
	stack := append([]int(nil), states...)
	for len(stack) > 0 {
		s := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if seen[s] {
			continue
		}
		seen[s] = true
		stack = append(stack, l.eps[s]...)
	}
	out := make([]int, 0, len(seen))
	for s := range seen {
		out = append(out, s)
	}
	sort.Ints(out)
	return out
}

func sortedLabels(trans map[pumlseqast.MessageKey]int) []pumlseqast.MessageKey {
	labels := make([]pumlseqast.MessageKey, 0, len(trans))
	for k := range trans {
		labels = append(labels, k)
	}
	slices.SortFunc(labels, pumlseqast.CompareMessageKey)
	return labels
}
