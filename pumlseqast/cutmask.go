package pumlseqast

import (
	"errors"

	"github.com/Kuniwak/puml-tools/pumlseq"
	"github.com/Kuniwak/puml-tools/sets"
)

var ErrNoMatch = errors.New("no matching group found")

func Cut(d *Diagram, match pumlseq.MatchFunc, name string) (*Diagram, error) {
	kept := collectMatchingGroups(d.Nodes, match)
	if len(kept) == 0 {
		return nil, ErrNoMatch
	}
	return CutBlocks(d, kept, name), nil
}

func CutBlocks(d *Diagram, kept []*Block, name string) *Diagram {
	keptSet := blockPointerSet(kept)

	refSet := sets.New()
	for _, b := range kept {
		gatherReferenced(b.Body, refSet)
		for _, e := range b.Elses {
			gatherReferenced(e.Body, refSet)
		}
	}

	var out []Node
	if start := findStart(d); start != nil {
		out = append(out, newStart(name, start.Term))
	}
	collectReferencedDecls(d.Nodes, keptSet, refSet, &out)
	for _, b := range kept {
		out = append(out, b)
	}
	if end := findEnd(d); end != nil {
		out = append(out, end)
	}
	return &Diagram{Nodes: out}
}

func Mask(d *Diagram, match pumlseq.MatchFunc, name string) (*Diagram, error) {
	matched := collectMatchingGroups(d.Nodes, match)
	if len(matched) == 0 {
		return nil, ErrNoMatch
	}
	matchedSet := blockPointerSet(matched)

	remaining := removeBlocks(d.Nodes, matchedSet)
	refSet := sets.New()
	gatherReferenced(remaining, refSet)
	return &Diagram{Nodes: rebuildMask(remaining, refSet, name)}, nil
}

func OutermostGroups(d *Diagram) []*Block {
	return collectMatchingGroups(d.Nodes, func(string) bool { return true })
}

func collectMatchingGroups(nodes []Node, match pumlseq.MatchFunc) []*Block {
	var out []*Block
	for _, n := range nodes {
		b, ok := n.(*Block)
		if !ok {
			continue
		}
		if b.Keyword == "group" && match(b.Label) {
			out = append(out, b)
			continue
		}
		out = append(out, collectMatchingGroups(b.Body, match)...)
		for _, e := range b.Elses {
			out = append(out, collectMatchingGroups(e.Body, match)...)
		}
	}
	return out
}

func gatherReferenced(nodes []Node, set sets.Set) {
	for _, n := range nodes {
		switch v := n.(type) {
		case *Message:
			set.Add(v.Sender)
			set.Add(v.Receiver)
		case *Activation:
			if v.Target != "" {
				set.Add(v.Target)
			}
		case *Note:
			for _, t := range v.Targets {
				set.Add(t)
			}
		case *Ref:
			for _, t := range v.Targets {
				set.Add(t)
			}
		case *Block:
			gatherReferenced(v.Body, set)
			for _, e := range v.Elses {
				gatherReferenced(e.Body, set)
			}
		}
	}
}

func collectReferencedDecls(nodes []Node, keptSet map[*Block]bool, refSet sets.Set, out *[]Node) {
	for _, n := range nodes {
		switch v := n.(type) {
		case *Participant:
			if refSet.Contains(v.Name) {
				*out = append(*out, v)
			}
		case *Block:
			if keptSet[v] {
				continue
			}
			collectReferencedDecls(v.Body, keptSet, refSet, out)
			for _, e := range v.Elses {
				collectReferencedDecls(e.Body, keptSet, refSet, out)
			}
		}
	}
}

func removeBlocks(nodes []Node, matchedSet map[*Block]bool) []Node {
	var out []Node
	for _, n := range nodes {
		b, ok := n.(*Block)
		if !ok {
			out = append(out, n)
			continue
		}
		if matchedSet[b] {
			continue
		}
		nb := *b
		nb.Body = removeBlocks(b.Body, matchedSet)
		if b.Elses != nil {
			nb.Elses = make([]ElseClause, len(b.Elses))
			for i, e := range b.Elses {
				ne := e
				ne.Body = removeBlocks(e.Body, matchedSet)
				nb.Elses[i] = ne
			}
		}
		out = append(out, &nb)
	}
	return out
}

func rebuildMask(nodes []Node, refSet sets.Set, name string) []Node {
	var out []Node
	for _, n := range nodes {
		switch v := n.(type) {
		case *Start:
			out = append(out, newStart(name, v.Term))
		case *Participant:
			if refSet.Contains(v.Name) {
				out = append(out, v)
			}
		case *Block:
			nb := *v
			nb.Body = rebuildMask(v.Body, refSet, name)
			if v.Elses != nil {
				nb.Elses = make([]ElseClause, len(v.Elses))
				for i, e := range v.Elses {
					ne := e
					ne.Body = rebuildMask(e.Body, refSet, name)
					nb.Elses[i] = ne
				}
			}
			out = append(out, &nb)
		default:
			out = append(out, n)
		}
	}
	return out
}

func newStart(name, term string) *Start {
	return &Start{Text: startumlHeader(name), Name: name, Term: term}
}

func startumlHeader(name string) string {
	if name == "" {
		return "@startuml"
	}
	return "@startuml " + name
}

func findStart(d *Diagram) *Start {
	for _, n := range d.Nodes {
		if s, ok := n.(*Start); ok {
			return s
		}
	}
	return nil
}

func findEnd(d *Diagram) *End {
	for _, n := range d.Nodes {
		if e, ok := n.(*End); ok {
			return e
		}
	}
	return nil
}

func blockPointerSet(bs []*Block) map[*Block]bool {
	m := make(map[*Block]bool, len(bs))
	for _, b := range bs {
		m[b] = true
	}
	return m
}
