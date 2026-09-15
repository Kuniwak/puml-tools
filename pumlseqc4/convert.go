package pumlseqc4

import (
	"github.com/Kuniwak/puml-tools/c4ast"
	"github.com/Kuniwak/puml-tools/pumlseqast"
)

func FromSeq(d *pumlseqast.Diagram) *c4ast.Diagram {
	return FromSeqs([]*pumlseqast.Diagram{d})
}

func FromSeqs(ds []*pumlseqast.Diagram) *c4ast.Diagram {
	out := &c4ast.Diagram{
		Elements:   make(map[c4ast.ElementID]c4ast.Element),
		Boundaries: make(map[c4ast.ElementID]c4ast.Boundary),
	}

	seenEl := make(map[c4ast.ElementID]bool)
	seenRel := make(map[[3]string]bool)
	for _, d := range ds {
		order, msgs := extractSeq(d)

		for _, p := range order {
			eid := c4ast.ElementID(p.ID)
			if seenEl[eid] {
				continue
			}
			seenEl[eid] = true
			out.Elements[eid] = c4ast.Element{ID: eid, Kind: p.Kind, Label: p.Label}
			out.Order = append(out.Order, eid)
		}

		for _, m := range msgs {
			if m.Src == m.Dst {
				continue
			}
			key := [3]string{m.Src, m.Dst, m.Label}
			if seenRel[key] {
				continue
			}
			seenRel[key] = true
			out.Rels = append(out.Rels, c4ast.Rel{From: c4ast.ElementID(m.Src), To: c4ast.ElementID(m.Dst), Label: m.Label})
		}
	}

	return out
}
