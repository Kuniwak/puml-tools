package c4ast

import (
	"errors"
	"fmt"
)

func Fold(d *Diagram, ids []string, newID, newDesc string) (*Diagram, error) {
	if len(ids) == 0 {
		return nil, errors.New("fold requires at least one element id")
	}

	mergeSet := make(map[ElementID]bool, len(ids))
	for _, id := range ids {
		eid := ElementID(id)
		if _, ok := d.Elements[eid]; !ok {
			return nil, fmt.Errorf("element not found: %s", id)
		}
		mergeSet[eid] = true
	}

	nid := ElementID(newID)
	if _, ok := d.Elements[nid]; ok && !mergeSet[nid] {
		return nil, fmt.Errorf("new-id %q conflicts with an existing element", newID)
	}

	label := newDesc
	if label == "" {
		label = newID
	}
	newEl := Element{ID: nid, Kind: "System", Label: label}

	out := &Diagram{
		Name:     d.Name,
		Elements: make(map[ElementID]Element, len(d.Elements)),
	}

	emitted := false
	for _, id := range d.Order {
		if mergeSet[id] {
			if !emitted {
				out.Order = append(out.Order, nid)
				out.Elements[nid] = newEl
				emitted = true
			}
			continue
		}
		out.Order = append(out.Order, id)
		el := d.Elements[id]
		el.Boundary = ""
		out.Elements[id] = el
	}

	remap := func(id ElementID) ElementID {
		if mergeSet[id] {
			return nid
		}
		return id
	}
	seen := make(map[[3]string]bool)
	for _, r := range d.Rels {
		from, to := remap(r.From), remap(r.To)
		if from == to {
			continue
		}
		key := [3]string{string(from), string(to), r.Label}
		if seen[key] {
			continue
		}
		seen[key] = true
		out.Rels = append(out.Rels, Rel{From: from, To: to, Label: r.Label})
	}

	return out, nil
}
