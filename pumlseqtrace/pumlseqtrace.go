package pumlseqtrace

import (
	"slices"
	"strings"

	"github.com/Kuniwak/puml-tools/pumlseqast"
)

type Options struct {
	LoopUnroll int
}

func Traces(d *pumlseqast.Diagram, opt Options) [][]pumlseqast.MessageKey {
	var out [][]pumlseqast.MessageKey
	Walk(d, opt, func(tr []pumlseqast.MessageKey) bool {
		out = append(out, tr)
		return true
	})
	return out
}

func Walk(d *pumlseqast.Diagram, opt Options, yield func(trace []pumlseqast.MessageKey) bool) {
	walkSeq(d.Nodes, opt, nil, func(tr []pumlseqast.MessageKey, stopped bool) bool {
		return yield(slices.Clone(tr))
	})
}

func walkSeq(nodes []pumlseqast.Node, opt Options, prefix []pumlseqast.MessageKey, k func(trace []pumlseqast.MessageKey, stopped bool) bool) bool {
	if len(nodes) == 0 {
		return k(prefix, false)
	}
	return walkNode(nodes[0], opt, prefix, func(tr []pumlseqast.MessageKey, stopped bool) bool {
		if stopped {
			return k(tr, true)
		}
		return walkSeq(nodes[1:], opt, tr, k)
	})
}

func walkNode(n pumlseqast.Node, opt Options, prefix []pumlseqast.MessageKey, k func(trace []pumlseqast.MessageKey, stopped bool) bool) bool {
	switch v := n.(type) {
	case *pumlseqast.Message:
		return k(append(prefix, v.Key()), false)
	case *pumlseqast.Block:
		return walkBlock(v, opt, prefix, k)
	case *pumlseqast.Ref:
		panic("pumlseqtrace: unexpanded ref over " + strings.Join(v.Targets, ", ") +
			" (call pumlseqast.ValidateNoRef first)")
	default:
		return k(prefix, false)
	}
}

func walkBlock(b *pumlseqast.Block, opt Options, prefix []pumlseqast.MessageKey, k func(trace []pumlseqast.MessageKey, stopped bool) bool) bool {
	if b.Keyword == "par" {
		return walkPar(branchBodies(b), opt, prefix, k)
	}
	if b.Keyword == "loop" {
		return walkLoop(b.Body, opt, prefix, k)
	}

	bodies := branchBodies(b)
	for _, body := range bodies {

		inner := k
		if b.Keyword == "break" {
			inner = func(tr []pumlseqast.MessageKey, _ bool) bool { return k(tr, true) }
		}
		if !walkSeq(body, opt, prefix, inner) {
			return false
		}
	}
	if b.Keyword == "opt" {
		return k(prefix, false)
	}
	return true
}

func walkLoop(body []pumlseqast.Node, opt Options, prefix []pumlseqast.MessageKey, k func(trace []pumlseqast.MessageKey, stopped bool) bool) bool {
	if opt.LoopUnroll <= 0 {
		return k(prefix, false)
	}
	return walkTurns(body, opt, 1, prefix, k)
}

func walkTurns(body []pumlseqast.Node, opt Options, turn int, prefix []pumlseqast.MessageKey, k func(trace []pumlseqast.MessageKey, stopped bool) bool) bool {
	return walkSeq(body, opt, prefix, func(tr []pumlseqast.MessageKey, stopped bool) bool {
		if stopped {
			return k(tr, true)
		}
		if !k(tr, false) {
			return false
		}
		if turn >= opt.LoopUnroll {
			return true
		}
		return walkTurns(body, opt, turn+1, slices.Clone(tr), k)
	})
}

func walkPar(bodies [][]pumlseqast.Node, opt Options, prefix []pumlseqast.MessageKey, k func(trace []pumlseqast.MessageKey, stopped bool) bool) bool {

	operands := make([][]partial, len(bodies))
	for i, body := range bodies {
		walkSeq(body, opt, nil, func(tr []pumlseqast.MessageKey, stopped bool) bool {
			operands[i] = append(operands[i], partial{Events: slices.Clone(tr), Stopped: stopped})
			return true
		})
	}
	return walkCombinations(operands, make([]partial, 0, len(operands)), prefix, k)
}

type partial struct {
	Events  []pumlseqast.MessageKey
	Stopped bool
}

func walkCombinations(rest [][]partial, chosen []partial, prefix []pumlseqast.MessageKey, k func(trace []pumlseqast.MessageKey, stopped bool) bool) bool {
	if len(rest) == 0 {
		return walkMerges(slices.Clone(chosen), prefix, k)
	}
	for _, p := range rest[0] {
		if !walkCombinations(rest[1:], append(chosen, p), prefix, k) {
			return false
		}
	}
	return true
}

func walkMerges(queues []partial, prefix []pumlseqast.MessageKey, k func(trace []pumlseqast.MessageKey, stopped bool) bool) bool {
	empty := true
	for _, q := range queues {
		if q.Stopped && len(q.Events) == 0 {
			return k(prefix, true)
		}
		if len(q.Events) > 0 {
			empty = false
		}
	}
	if empty {
		return k(prefix, false)
	}
	for i, q := range queues {
		if len(q.Events) == 0 {
			continue
		}

		next := slices.Clone(queues)
		next[i].Events = q.Events[1:]
		if !walkMerges(next, append(slices.Clone(prefix), q.Events[0]), k) {
			return false
		}
	}
	return true
}

func branchBodies(b *pumlseqast.Block) [][]pumlseqast.Node {
	bodies := [][]pumlseqast.Node{b.Body}
	for _, e := range b.Elses {
		bodies = append(bodies, e.Body)
	}
	return bodies
}
