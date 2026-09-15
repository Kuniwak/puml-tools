package pumlseqc4

import (
	"fmt"
	"strings"
)

func (d *Diff) ReportEq() string {
	var sb strings.Builder
	writeNodeSection(&sb, "participants only in SEQ (missing in C4)", d.SeqOnlyNodes)
	writeNodeSection(&sb, "elements only in C4 (missing in SEQ)", d.C4OnlyNodes)
	writeEdgeSection(&sb, "messages only in SEQ (missing in C4)", d.SeqOnlyEdges)
	writeEdgeSection(&sb, "relationships only in C4 (missing in SEQ)", d.C4OnlyEdges)
	return sb.String()
}

func (d *Diff) ReportSubset() string {
	var sb strings.Builder
	writeNodeSection(&sb, "elements only in C4 (not in SEQ)", d.C4OnlyNodes)
	writeEdgeSection(&sb, "relationships only in C4 (not in SEQ)", d.C4OnlyEdges)
	return sb.String()
}

func (d *Diff) ReportSuperset() string {
	var sb strings.Builder
	writeNodeSection(&sb, "participants only in SEQ (not in C4)", d.SeqOnlyNodes)
	writeEdgeSection(&sb, "messages only in SEQ (not in C4)", d.SeqOnlyEdges)
	return sb.String()
}

func writeNodeSection(sb *strings.Builder, title string, nodes []Node) {
	if len(nodes) == 0 {
		return
	}
	fmt.Fprintf(sb, "%s: %d\n", title, len(nodes))
	for _, n := range nodes {
		fmt.Fprintf(sb, "  - %s (%q)\n", n.ID, n.Label)
	}
}

func writeEdgeSection(sb *strings.Builder, title string, edges []Edge) {
	if len(edges) == 0 {
		return
	}
	fmt.Fprintf(sb, "%s: %d\n", title, len(edges))
	for _, e := range edges {
		fmt.Fprintf(sb, "  - %s -> %s\n", e.From, e.To)
	}
}
