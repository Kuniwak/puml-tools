package pumlseqsubset

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
)

type Pair struct {
	Sender   string
	Receiver string
}

type Seq struct {
	Items []Item
}

type Item struct {
	Pair     Pair
	Branches []Seq
}

func NewSeq(pairs ...Pair) Seq {
	items := make([]Item, 0, len(pairs))
	for _, p := range pairs {
		items = append(items, Item{Pair: p})
	}
	return Seq{Items: items}
}

func NewPar(branches ...Seq) Item {
	if branches == nil {
		branches = []Seq{}
	}
	return Item{Branches: branches}
}

func Pairs(s Seq) []Pair {
	var out []Pair
	appendPairs(s, &out)
	return out
}

func appendPairs(s Seq, out *[]Pair) {
	for _, it := range s.Items {
		if it.Branches == nil {
			*out = append(*out, it.Pair)
			continue
		}
		for _, b := range it.Branches {
			appendPairs(b, out)
		}
	}
}

type Mode string

const (
	ModeExact Mode = "exact"

	ModeAllowSubseq Mode = "allow-subseq"
)

func ParseMode(s string) (Mode, error) {
	switch Mode(s) {
	case ModeExact, ModeAllowSubseq:
		return Mode(s), nil
	default:
		return "", fmt.Errorf("invalid mode %q: must be %q or %q", s, ModeExact, ModeAllowSubseq)
	}
}

type Side string

const (
	SideSub   Side = "sub"
	SideSuper Side = "super"
)

type Residual struct {
	Side  Side
	Index int
	Pair  Pair
}

func Residuals(mode Mode, sub, super Seq) []Residual {
	if mode == ModeAllowSubseq {

		subOnly, _ := residuals(sub, super, 0)
		return appendResiduals(nil, SideSub, sub, subOnly)
	}

	subOnly, superOnly := residuals(sub, super, 1)
	out := appendResiduals(nil, SideSub, sub, subOnly)
	return appendResiduals(out, SideSuper, super, superOnly)
}

func Record(mode Mode, r Residual) []string {
	switch mode {
	case ModeAllowSubseq:
		return []string{r.Pair.Sender, r.Pair.Receiver}
	case ModeExact:
		return []string{string(r.Side), r.Pair.Sender, r.Pair.Receiver}
	default:
		panic(fmt.Sprintf("pumlseqsubset: invalid mode %q", mode))
	}
}

func appendResiduals(out []Residual, side Side, s Seq, idxs []int) []Residual {
	if len(idxs) == 0 {
		return out
	}
	pairs := Pairs(s)
	for _, i := range idxs {
		out = append(out, Residual{Side: side, Index: i, Pair: pairs[i]})
	}
	return out
}

func residuals(sub, super Seq, skipSuperCost int) ([]int, []int) {
	m := newMatcher(sub, super, skipSuperCost)

	var subOnly, superOnly []int
	sp, pp := enter(m.sub, 0), enter(m.super, 0)
	cost := m.minCost(sp, pp)
	for !finished(m.sub, sp) || !finished(m.super, pp) {
		mv, ok := m.step(sp, pp, cost)
		if !ok {

			panic("pumlseqsubset: no optimal transition")
		}
		if mv.DropSub >= 0 {
			subOnly = append(subOnly, mv.DropSub)
			cost--
		}
		if mv.SkipSuper >= 0 {
			superOnly = append(superOnly, mv.SkipSuper)
			cost -= skipSuperCost
		}
		sp, pp = mv.Sub, mv.Super
	}
	slices.Sort(subOnly)
	slices.Sort(superOnly)
	return subOnly, superOnly
}

type flow struct {
	Items []flowItem
}

type flowItem struct {
	Index    int
	Pair     Pair
	Branches []flow
}

func newFlow(s Seq) flow {
	next := 0
	return buildFlow(s, &next)
}

func buildFlow(s Seq, next *int) flow {
	items := make([]flowItem, 0, len(s.Items))
	for _, it := range s.Items {
		if it.Branches == nil {
			items = append(items, flowItem{Index: *next, Pair: it.Pair})
			*next++
			continue
		}
		branches := make([]flow, 0, len(it.Branches))
		for _, b := range it.Branches {
			branches = append(branches, buildFlow(b, next))
		}
		items = append(items, flowItem{Index: -1, Branches: branches})
	}
	return flow{Items: items}
}

type pos struct {
	Index    int
	Branches []pos
}

func finished(f flow, p pos) bool { return p.Index >= len(f.Items) }

func enter(f flow, i int) pos {
	for i < len(f.Items) {
		it := f.Items[i]
		if it.Branches == nil {
			return pos{Index: i}
		}
		branches := make([]pos, len(it.Branches))
		allFinished := true
		for k, b := range it.Branches {
			branches[k] = enter(b, 0)
			if !finished(b, branches[k]) {
				allFinished = false
			}
		}
		if !allFinished {
			return pos{Index: i, Branches: branches}
		}
		i++
	}
	return pos{Index: i}
}

func enabled(f flow, p pos, out *[]int) {
	if finished(f, p) {
		return
	}
	it := f.Items[p.Index]
	if it.Branches == nil {
		*out = append(*out, it.Index)
		return
	}
	for k, b := range it.Branches {
		enabled(b, p.Branches[k], out)
	}
}

func advance(f flow, p pos, idx int) (pos, bool) {
	if finished(f, p) {
		return p, false
	}
	it := f.Items[p.Index]
	if it.Branches == nil {
		if it.Index != idx {
			return p, false
		}
		return enter(f, p.Index+1), true
	}
	for k, b := range it.Branches {
		next, ok := advance(b, p.Branches[k], idx)
		if !ok {
			continue
		}
		branches := make([]pos, len(p.Branches))
		copy(branches, p.Branches)
		branches[k] = next
		for j, bp := range branches {
			if !finished(it.Branches[j], bp) {
				return pos{Index: p.Index, Branches: branches}, true
			}
		}
		return enter(f, p.Index+1), true
	}
	return p, false
}

type state struct {
	Sub   string
	Super string
}

type matcher struct {
	sub           flow
	super         flow
	subPairs      []Pair
	superPairs    []Pair
	skipSuperCost int
	memo          map[state]int
}

func newMatcher(sub, super Seq, skipSuperCost int) *matcher {
	return &matcher{
		sub:           newFlow(sub),
		super:         newFlow(super),
		subPairs:      Pairs(sub),
		superPairs:    Pairs(super),
		skipSuperCost: skipSuperCost,
		memo:          map[state]int{},
	}
}

type move struct {
	Sub       pos
	Super     pos
	DropSub   int
	SkipSuper int
}

func (m *matcher) minCost(sp, pp pos) int {
	if finished(m.sub, sp) {

		if m.skipSuperCost == 0 || finished(m.super, pp) {
			return 0
		}
	}
	key := state{Sub: posKey(sp), Super: posKey(pp)}
	if c, ok := m.memo[key]; ok {
		return c
	}

	var subEvents, superEvents []int
	enabled(m.sub, sp, &subEvents)
	enabled(m.super, pp, &superEvents)

	best := math.MaxInt / 2
	for _, e := range subEvents {
		next, _ := advance(m.sub, sp, e)
		if c := 1 + m.minCost(next, pp); c < best {
			best = c
		}
	}
	for _, f := range superEvents {
		nextSuper, _ := advance(m.super, pp, f)
		if c := m.skipSuperCost + m.minCost(sp, nextSuper); c < best {
			best = c
		}
		for _, e := range subEvents {
			if m.subPairs[e] != m.superPairs[f] {
				continue
			}
			nextSub, _ := advance(m.sub, sp, e)
			if c := m.minCost(nextSub, nextSuper); c < best {
				best = c
			}
		}
	}
	m.memo[key] = best
	return best
}

func (m *matcher) step(sp, pp pos, cost int) (move, bool) {
	var subEvents, superEvents []int
	enabled(m.sub, sp, &subEvents)
	enabled(m.super, pp, &superEvents)

	for _, f := range superEvents {
		nextSuper, _ := advance(m.super, pp, f)
		for _, e := range subEvents {
			if m.subPairs[e] != m.superPairs[f] {
				continue
			}
			nextSub, _ := advance(m.sub, sp, e)
			if m.minCost(nextSub, nextSuper) == cost {
				return move{Sub: nextSub, Super: nextSuper, DropSub: -1, SkipSuper: -1}, true
			}
		}
	}
	for _, e := range subEvents {
		nextSub, _ := advance(m.sub, sp, e)
		if 1+m.minCost(nextSub, pp) == cost {
			return move{Sub: nextSub, Super: pp, DropSub: e, SkipSuper: -1}, true
		}
	}
	for _, f := range superEvents {
		nextSuper, _ := advance(m.super, pp, f)
		if m.skipSuperCost+m.minCost(sp, nextSuper) == cost {
			return move{Sub: sp, Super: nextSuper, DropSub: -1, SkipSuper: f}, true
		}
	}
	return move{DropSub: -1, SkipSuper: -1}, false
}

func posKey(p pos) string {
	if len(p.Branches) == 0 {
		return strconv.Itoa(p.Index)
	}
	var b strings.Builder
	b.WriteString(strconv.Itoa(p.Index))
	b.WriteByte('(')
	for i, c := range p.Branches {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(posKey(c))
	}
	b.WriteByte(')')
	return b.String()
}
