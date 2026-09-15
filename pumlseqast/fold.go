package pumlseqast

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Kuniwak/puml-tools/lex"
	"github.com/Kuniwak/puml-tools/sets"
)

func Fold(d *Diagram, ids []string, newID, newDesc string) (*Diagram, error) {
	if len(ids) == 0 {
		return nil, errors.New("fold requires at least one participant id")
	}
	mergeSet := sets.New(ids...)

	present := DeclaredNames(d.Nodes)
	present.AddAll(ReferencedNames(d.Nodes))
	for _, id := range ids {
		if !present.Contains(id) {
			return nil, fmt.Errorf("participant not found: %s", id)
		}
	}
	if present.Contains(newID) && !mergeSet.Contains(newID) {
		return nil, fmt.Errorf("new-id %q conflicts with an existing participant", newID)
	}

	newDecl := newDeclaration(newID, newDesc)
	emitted := false
	var out []Node
	for _, n := range d.Nodes {
		if p, ok := n.(*Participant); ok && mergeSet.Contains(p.Name) {
			if !emitted {
				out = append(out, newDecl)
				emitted = true
			}
			continue
		}
		if !emitted && touchesNode(n, mergeSet) {
			out = append(out, newDecl)
			emitted = true
		}
		out = append(out, transformNode(n, mergeSet, newID))
	}
	return &Diagram{Nodes: out}, nil
}

func touchesNode(n Node, set sets.Set) bool {
	switch v := n.(type) {
	case *Participant:
		return set.Contains(v.Name)
	case *Message:
		return set.Contains(lex.Unquote(v.LeftRaw)) || set.Contains(lex.Unquote(v.RightRaw))
	case *Activation:
		return set.Contains(v.Target)
	case *Note:
		for _, t := range v.Targets {
			if set.Contains(t) {
				return true
			}
		}
		return false
	case *Ref:
		for _, t := range v.Targets {
			if set.Contains(t) {
				return true
			}
		}
		return false
	case *Block:
		for _, c := range v.Body {
			if touchesNode(c, set) {
				return true
			}
		}
		for _, e := range v.Elses {
			for _, c := range e.Body {
				if touchesNode(c, set) {
					return true
				}
			}
		}
		return false
	default:
		return false
	}
}

func transformNodes(nodes []Node, set sets.Set, newID string) []Node {
	var out []Node
	for _, n := range nodes {
		if p, ok := n.(*Participant); ok && set.Contains(p.Name) {
			continue
		}
		out = append(out, transformNode(n, set, newID))
	}
	return out
}

func transformNode(n Node, set sets.Set, newID string) Node {
	switch v := n.(type) {
	case *Message:
		return renameMessage(v, set, newID)
	case *Activation:
		return renameActivation(v, set, newID)
	case *Note:
		return renameNote(v, set, newID)
	case *Ref:
		return renameRef(v, set, newID)
	case *Block:
		nb := *v
		nb.Body = transformNodes(v.Body, set, newID)
		if v.Elses != nil {
			nb.Elses = make([]ElseClause, len(v.Elses))
			for i, e := range v.Elses {
				ne := e
				ne.Body = transformNodes(e.Body, set, newID)
				nb.Elses[i] = ne
			}
		}
		return &nb
	default:
		return n
	}
}

func renameMessage(m *Message, set sets.Set, newID string) *Message {
	leftMerged := set.Contains(lex.Unquote(m.LeftRaw))
	rightMerged := set.Contains(lex.Unquote(m.RightRaw))
	if !leftMerged && !rightMerged {
		return m
	}
	nm := *m
	if leftMerged {
		nm.LeftRaw = renderID(newID)
	}
	if rightMerged {
		nm.RightRaw = renderID(newID)
	}
	if set.Contains(nm.Sender) {
		nm.Sender = newID
	}
	if set.Contains(nm.Receiver) {
		nm.Receiver = newID
	}
	return &nm
}

func renameActivation(a *Activation, set sets.Set, newID string) Node {
	if !set.Contains(a.Target) {
		return a
	}
	na := *a
	na.TargetRaw = renderID(newID)
	na.Target = newID
	return &na
}

func renameNote(n *Note, set sets.Set, newID string) Node {
	renamed, raw, changed := renameTargets(n.Targets, set, newID)
	if !changed {
		return n
	}
	nn := *n
	nn.Targets = renamed
	nn.TargetsRaw = raw
	return &nn
}

func renameRef(n *Ref, set sets.Set, newID string) Node {
	renamed, raw, changed := renameTargets(n.Targets, set, newID)
	if !changed {
		return n
	}
	nn := *n
	nn.Targets = renamed
	nn.TargetsRaw = raw
	return &nn
}

func renameTargets(targets []string, set sets.Set, newID string) (renamed []string, raw string, changed bool) {
	if len(targets) == 0 {
		return nil, "", false
	}
	seen := sets.New()
	for _, t := range targets {
		nt := t
		if set.Contains(t) {
			nt = newID
			changed = true
		}
		if seen.Contains(nt) {
			changed = true
			continue
		}
		seen.Add(nt)
		renamed = append(renamed, nt)
	}
	if !changed {
		return nil, "", false
	}
	parts := make([]string, len(renamed))
	for i, t := range renamed {
		parts[i] = renderID(t)
	}
	return renamed, strings.Join(parts, ", "), true
}

func newDeclaration(newID, newDesc string) *Participant {
	content := "participant " + renderID(newID)
	if newDesc != "" {
		content = "participant \"" + newDesc + "\" as " + newID
	}
	return &Participant{Text: content, Keyword: "participant", Name: newID, Term: "\n"}
}

func renderID(id string) string {
	if strings.ContainsAny(id, " \t") {
		return "\"" + id + "\""
	}
	return id
}
