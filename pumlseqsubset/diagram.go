package pumlseqsubset

import (
	"strings"

	"github.com/Kuniwak/puml-tools/pumlseqast"
)

const ParKeyword = "par"

func SeqOfDiagram(d *pumlseqast.Diagram) Seq {
	return SeqOfNodes(d.Nodes)
}

func SeqOfNodes(nodes []pumlseqast.Node) Seq {

	items := []Item{}
	for _, n := range nodes {
		switch v := n.(type) {
		case *pumlseqast.Message:
			items = append(items, Item{Pair: Pair{Sender: v.Sender, Receiver: v.Receiver}})
		case *pumlseqast.Ref:
			panic("pumlseqsubset: unexpanded ref over " + strings.Join(v.Targets, ", ") +
				" (call pumlseqast.ValidateNoRef first)")
		case *pumlseqast.Block:
			if v.Keyword == ParKeyword {
				items = append(items, NewPar(SeqsOfBlock(v)...))
				continue
			}
			for _, b := range SeqsOfBlock(v) {
				items = append(items, b.Items...)
			}
		}
	}
	return Seq{Items: items}
}

func SeqsOfBlock(b *pumlseqast.Block) []Seq {
	seqs := []Seq{SeqOfNodes(b.Body)}
	for _, e := range b.Elses {
		seqs = append(seqs, SeqOfNodes(e.Body))
	}
	return seqs
}
