package c4ast

import "fmt"

func BoundaryMembers(d *Diagram, boundary ElementID) ([]ElementID, error) {
	if _, ok := d.Boundaries[boundary]; !ok {
		return nil, fmt.Errorf("boundary not found: %s", boundary)
	}
	var members []ElementID
	for _, id := range d.Order {
		if withinBoundary(d, d.Elements[id].Boundary, boundary) {
			members = append(members, id)
		}
	}
	return members, nil
}

func withinBoundary(d *Diagram, start, target ElementID) bool {
	for cur := start; cur != ""; {
		if cur == target {
			return true
		}
		b, ok := d.Boundaries[cur]
		if !ok {
			return false
		}
		cur = b.Parent
	}
	return false
}

func (d *Diagram) EnclosingBoundary(id ElementID) ElementID {
	direct := d.Elements[id].Boundary
	if b, ok := d.Boundaries[id]; ok {
		direct = b.Parent
	}
	seen := map[ElementID]bool{id: true}
	for cur := direct; cur != ""; {
		b, ok := d.Boundaries[cur]
		if !ok || seen[cur] {
			return ""
		}
		seen[cur] = true
		cur = b.Parent
	}
	return direct
}
