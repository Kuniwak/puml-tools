package pumlseqast

import (
	"fmt"
	"sort"

	"github.com/Kuniwak/puml-tools/sets"
)

type Folding map[string]string

func (f Folding) Image(id string) string {
	if mapped, ok := f[id]; ok {
		return mapped
	}
	return id
}

func ValidateFolding(phi Folding, abstract, detail *Diagram) error {
	abstractParts := sets.New(Participants(abstract)...)
	for _, p := range Participants(detail) {
		image := phi.Image(p)
		if !abstractParts.Contains(image) {
			return fmt.Errorf("detail participant %q maps to %q, which is not a participant of the abstract diagram", p, image)
		}
	}
	return nil
}

func Participants(d *Diagram) []string {
	set := DeclaredNames(d.Nodes)
	set.AddAll(ReferencedNames(d.Nodes))
	return set.Sorted()
}

func FoldAll(d *Diagram, phi Folding) (*Diagram, error) {
	preimage := map[string][]string{}
	for det, abs := range phi {
		preimage[abs] = append(preimage[abs], det)
	}
	targets := make([]string, 0, len(preimage))
	for t := range preimage {
		targets = append(targets, t)
	}
	sort.Strings(targets)

	out := d
	for _, t := range targets {
		g := preimage[t]
		sort.Strings(g)
		if len(g) == 1 && g[0] == t {
			continue
		}
		folded, err := Fold(out, g, t, "")
		if err != nil {
			return nil, err
		}
		out = folded
	}
	return out, nil
}
