package pumlseqast

import "github.com/Kuniwak/puml-tools/pumlseq"

func PruneEmptyBoxes(d *Diagram) *Diagram {
	dropped := emptyBoxFrames(d.Nodes)
	return &Diagram{Nodes: DropRawLines(d.Nodes, func(nl NodeLine) bool { return dropped[nl] })}
}

func emptyBoxFrames(nodes []Node) map[NodeLine]bool {
	items := NodeLines(nodes)
	dropped := make(map[NodeLine]bool)
	for k := 0; k < len(items); k++ {
		if items[k].Line < 0 || !pumlseq.IsBoxStart(items[k].Text) {
			continue
		}
		end := -1
		enclosesDeclaration := false
		for j := k + 1; j < len(items); j++ {
			if items[j].Line >= 0 && pumlseq.IsBoxEnd(items[j].Text) {
				end = j
				break
			}

			if items[j].Line < 0 && len(DeclaredNames(nodes[items[j].Node:items[j].Node+1])) > 0 {
				enclosesDeclaration = true
			}
		}
		if end < 0 {
			continue
		}
		if !enclosesDeclaration {
			dropped[items[k]] = true
			dropped[items[end]] = true
		}
		k = end
	}
	return dropped
}
