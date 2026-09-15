package c4ast

import (
	"fmt"
	"sort"
	"strings"
)

type ElementID string

type Element struct {
	ID       ElementID
	Kind     string
	External bool
	Label    string
	Descr    string
	Boundary ElementID
}

type Boundary struct {
	ID     ElementID
	Kind   string
	Label  string
	Parent ElementID
}

type Rel struct {
	From  ElementID
	To    ElementID
	Label string
}

type Diagram struct {
	Name       string
	Order      []ElementID
	Elements   map[ElementID]Element
	Boundaries map[ElementID]Boundary
	Rels       []Rel
}

func (d *Diagram) String() string {
	var sb strings.Builder
	if d.Name != "" {
		fmt.Fprintf(&sb, "@startuml %s\n", d.Name)
	} else {
		sb.WriteString("@startuml\n")
	}
	fmt.Fprintf(&sb, "!include <C4/%s>\n\n", d.Include())

	sb.WriteString(d.renderChildren("", "  "))

	if len(d.Rels) > 0 {
		sb.WriteByte('\n')
		for _, rel := range d.Rels {
			fmt.Fprintf(&sb, "Rel(%s, %s, %q)\n", rel.From, rel.To, rel.Label)
		}
	}

	sb.WriteString("@enduml\n")
	return sb.String()
}

func (d *Diagram) Include() string {
	kinds := make([]string, 0, len(d.Order)+len(d.Boundaries))
	for _, id := range d.Order {
		kinds = append(kinds, d.Elements[id].Kind)
	}
	for _, id := range sortedBoundaryIDs(d.Boundaries) {
		kinds = append(kinds, d.Boundaries[id].Kind)
	}

	include := "C4_Context"
	for _, kind := range kinds {
		switch {
		case strings.HasPrefix(kind, "Component"):
			return "C4_Component"
		case strings.HasPrefix(kind, "Container"):
			include = "C4_Container"
		}
	}
	return include
}

func sortedBoundaryIDs(boundaries map[ElementID]Boundary) []ElementID {
	ids := make([]ElementID, 0, len(boundaries))
	for id := range boundaries {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func (d *Diagram) renderChildren(parent ElementID, indent string) string {
	var sb strings.Builder
	for _, c := range d.childrenOf(parent) {
		if b, ok := d.Boundaries[c]; ok {
			fmt.Fprintf(&sb, "%s(%s, %q) {\n", b.Kind, b.ID, b.Label)
			for _, line := range strings.SplitAfter(d.renderChildren(c, indent), "\n") {
				if line != "" {
					sb.WriteString(indent + line)
				}
			}
			sb.WriteString("}\n")
			continue
		}
		sb.WriteString(renderElement(d.Elements[c]))
		sb.WriteByte('\n')
	}
	return sb.String()
}

func (d *Diagram) childrenOf(parent ElementID) []ElementID {
	rank := make(map[ElementID]int, len(d.Boundaries)+len(d.Order))
	for i, id := range d.Order {
		for cur := d.EnclosingBoundary(id); cur != ""; cur = d.Boundaries[cur].Parent {
			if _, seen := rank[cur]; seen {
				break
			}
			rank[cur] = i
		}
	}

	var children []ElementID
	for i, id := range d.Order {
		if d.EnclosingBoundary(id) == parent {
			children = append(children, id)
			rank[id] = i
		}
	}
	var empty []ElementID
	for id, b := range d.Boundaries {
		if b.Parent != parent || d.EnclosingBoundary(id) != b.Parent {
			continue
		}
		if _, ok := rank[id]; ok {
			children = append(children, id)
			continue
		}
		empty = append(empty, id)
	}
	sort.Slice(children, func(i, j int) bool { return rank[children[i]] < rank[children[j]] })
	sort.Slice(empty, func(i, j int) bool { return empty[i] < empty[j] })
	return append(children, empty...)
}

func renderElement(e Element) string {
	macro := e.Kind
	if e.External {
		macro += "_Ext"
	}
	if e.Descr != "" {
		return fmt.Sprintf("%s(%s, %q, %q)", macro, e.ID, e.Label, e.Descr)
	}
	return fmt.Sprintf("%s(%s, %q)", macro, e.ID, e.Label)
}
