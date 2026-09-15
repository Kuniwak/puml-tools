package pumlseqc4

import (
	"fmt"
	"sort"

	"github.com/Kuniwak/puml-tools/c4ast"
	"github.com/Kuniwak/puml-tools/pumlseqast"
)

type Node struct {
	ID    string
	Label string
}

type Edge struct {
	From string
	To   string
}

type Diff struct {
	SeqOnlyNodes []Node
	C4OnlyNodes  []Node
	SeqOnlyEdges []Edge
	C4OnlyEdges  []Edge
}

func Compare(seqD *pumlseqast.Diagram, c4D *c4ast.Diagram) *Diff {
	order, msgs := extractSeq(seqD)
	return compareModel(order, msgs, c4D)
}

func compareModel(order []participant, msgs []message, c4D *c4ast.Diagram) *Diff {
	seqNodes := make(map[string]bool, len(order))
	for _, p := range order {
		seqNodes[p.ID] = true
	}
	c4Nodes := make(map[string]bool, len(c4D.Order))
	for _, id := range c4D.Order {
		c4Nodes[string(id)] = true
	}

	seqEdges := make(map[Edge]bool)
	for _, m := range msgs {
		if m.Src == m.Dst {
			continue
		}
		seqEdges[Edge{From: m.Src, To: m.Dst}] = true
	}
	c4Edges := make(map[Edge]bool)
	for _, r := range c4D.Rels {
		if r.From == r.To {
			continue
		}
		c4Edges[Edge{From: string(r.From), To: string(r.To)}] = true
	}

	d := &Diff{}
	for _, p := range order {
		if !c4Nodes[p.ID] {
			d.SeqOnlyNodes = append(d.SeqOnlyNodes, Node{ID: p.ID, Label: p.Label})
		}
	}
	for _, id := range c4D.Order {
		if !seqNodes[string(id)] {
			d.C4OnlyNodes = append(d.C4OnlyNodes, Node{ID: string(id), Label: c4D.Elements[id].Label})
		}
	}
	for e := range seqEdges {
		if !c4Edges[e] {
			d.SeqOnlyEdges = append(d.SeqOnlyEdges, e)
		}
	}
	for e := range c4Edges {
		if !seqEdges[e] {
			d.C4OnlyEdges = append(d.C4OnlyEdges, e)
		}
	}

	sortNodes(d.SeqOnlyNodes)
	sortNodes(d.C4OnlyNodes)
	sortEdges(d.SeqOnlyEdges)
	sortEdges(d.C4OnlyEdges)
	return d
}

func (d *Diff) IsSubset() bool {
	return len(d.C4OnlyNodes) == 0 && len(d.C4OnlyEdges) == 0
}

func (d *Diff) IsSuperset() bool {
	return len(d.SeqOnlyNodes) == 0 && len(d.SeqOnlyEdges) == 0
}

func (d *Diff) IsEqual() bool {
	return len(d.SeqOnlyNodes) == 0 && len(d.C4OnlyNodes) == 0 &&
		len(d.SeqOnlyEdges) == 0 && len(d.C4OnlyEdges) == 0
}

func sortNodes(ns []Node) {
	sort.Slice(ns, func(i, j int) bool { return ns[i].ID < ns[j].ID })
}

func sortEdges(es []Edge) {
	sort.Slice(es, func(i, j int) bool {
		if es[i].From != es[j].From {
			return es[i].From < es[j].From
		}
		return es[i].To < es[j].To
	})
}

type ParticipantConflict struct {
	ID         string
	FirstIndex int
	FirstLabel string
	FirstKind  string
	Index      int
	Label      string
	Kind       string
}

func (c *ParticipantConflict) Error() string {
	return fmt.Sprintf("participant %q conflicts between input #%d and input #%d: %s %q vs %s %q",
		c.ID, c.FirstIndex+1, c.Index+1, c.FirstKind, c.FirstLabel, c.Kind, c.Label)
}

func CompareSeqs(seqDs []*pumlseqast.Diagram, c4D *c4ast.Diagram) (*Diff, error) {
	order, msgs, err := mergeSeqs(seqDs)
	if err != nil {
		return nil, err
	}
	return compareModel(order, msgs, c4D), nil
}

func mergeSeqs(seqDs []*pumlseqast.Diagram) ([]participant, []message, error) {
	var order []participant
	var msgs []message
	first := make(map[string]participant)
	firstIndex := make(map[string]int)

	for i, d := range seqDs {
		ps, ms := extractSeq(d)
		for _, p := range ps {
			prev, seen := first[p.ID]
			if !seen {
				first[p.ID] = p
				firstIndex[p.ID] = i
				order = append(order, p)
				continue
			}
			if !prev.Declared || !p.Declared {
				continue
			}
			if prev.Label != p.Label || prev.Kind != p.Kind {
				return nil, nil, &ParticipantConflict{
					ID:         p.ID,
					FirstIndex: firstIndex[p.ID],
					FirstLabel: prev.Label,
					FirstKind:  prev.Kind,
					Index:      i,
					Label:      p.Label,
					Kind:       p.Kind,
				}
			}
		}
		msgs = append(msgs, ms...)
	}
	return order, msgs, nil
}
