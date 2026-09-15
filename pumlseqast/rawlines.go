package pumlseqast

import "strings"

type NodeLine struct {
	Node int
	Line int
	Text string
}

func NodeLines(nodes []Node) []NodeLine {
	var out []NodeLine
	for i, n := range nodes {
		r, ok := n.(*Raw)
		if !ok {
			out = append(out, NodeLine{Node: i, Line: -1})
			continue
		}
		for j, pl := range splitPhysLines(r.Text) {
			out = append(out, NodeLine{Node: i, Line: j, Text: pl.content})
		}
	}
	return out
}

func DropRawLines(nodes []Node, drop func(NodeLine) bool) []Node {
	out := make([]Node, 0, len(nodes))
	for i, n := range nodes {
		r, ok := n.(*Raw)
		if !ok {
			out = append(out, n)
			continue
		}
		var b strings.Builder
		for j, pl := range splitPhysLines(r.Text) {
			if drop(NodeLine{Node: i, Line: j, Text: pl.content}) {
				continue
			}
			b.WriteString(pl.content + pl.term)
		}
		if b.Len() == 0 {
			continue
		}
		out = append(out, &Raw{Text: b.String()})
	}
	return out
}
