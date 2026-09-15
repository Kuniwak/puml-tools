package pumlseqc4

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/Kuniwak/puml-tools/c4ast"
	"github.com/Kuniwak/puml-tools/pumlseq"
	"github.com/Kuniwak/puml-tools/pumlseqast"
)

const BoundaryKind = "System_Boundary"

func BoxOf(d *pumlseqast.Diagram) map[string]string {
	out := make(map[string]string)
	box := ""
	for _, item := range pumlseqast.NodeLines(d.Nodes) {
		if item.Line >= 0 {
			switch {
			case pumlseq.IsBoxEnd(item.Text):
				box = ""
			case pumlseq.IsBoxStart(item.Text):
				label, ok := pumlseq.BoxLabel(item.Text)
				if !ok {
					box = ""
					continue
				}
				box = NormalizeBoxLabel(label)
			}
			continue
		}
		if box == "" {
			continue
		}
		if p, ok := d.Nodes[item.Node].(*pumlseqast.Participant); ok {
			out[p.Name] = box
		}
	}
	return out
}

func BoundaryID(label string) c4ast.ElementID {
	var sb strings.Builder
	for _, r := range strings.ToLower(label) {
		switch {
		case r > 127:
			sum := sha1.Sum([]byte(label))
			return c4ast.ElementID("bdy_" + hex.EncodeToString(sum[:4]))
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			sb.WriteRune(r)
		default:
			sb.WriteByte('_')
		}
	}
	slug := strings.Trim(collapseUnderscores(sb.String()), "_")
	if slug == "" {
		sum := sha1.Sum([]byte(label))
		return c4ast.ElementID("bdy_" + hex.EncodeToString(sum[:4]))
	}
	return c4ast.ElementID(slug)
}

func collapseUnderscores(s string) string {
	for strings.Contains(s, "__") {
		s = strings.ReplaceAll(s, "__", "_")
	}
	return s
}

func FromSeqsWithBoxBoundaries(ds []*pumlseqast.Diagram) (*c4ast.Diagram, error) {
	out := FromSeqs(ds)

	boxOf := make(map[string]string)
	for _, d := range ds {
		for id, label := range BoxOf(d) {
			if prev, ok := boxOf[id]; ok && prev != label {
				return nil, fmt.Errorf("participant %q belongs to different boxes across diagrams: %q and %q", id, prev, label)
			}
			boxOf[id] = label
		}
	}

	for _, eid := range out.Order {
		label, ok := boxOf[string(eid)]
		if !ok {
			continue
		}
		bid := BoundaryID(label)
		if prev, ok := out.Boundaries[bid]; ok && prev.Label != label {
			return nil, fmt.Errorf("boxes %q and %q map to the same boundary id %q", prev.Label, label, bid)
		}
		out.Boundaries[bid] = c4ast.Boundary{ID: bid, Kind: BoundaryKind, Label: label}
		e := out.Elements[eid]
		e.Boundary = bid
		out.Elements[eid] = e
	}
	return out, nil
}
