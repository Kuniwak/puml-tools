package pumlseqast

import (
	"strings"

	"github.com/Kuniwak/puml-tools/lines"
)

func Prune(d *Diagram) *Diagram {
	return &Diagram{Nodes: pruneNodes(d.Nodes)}
}

func pruneNodes(nodes []Node) []Node {
	var out []Node
	for _, n := range nodes {
		if b, ok := n.(*Block); ok {
			out = append(out, pruneBlock(b)...)
			continue
		}
		out = append(out, n)
	}
	return out
}

type compartment struct {
	first   bool
	sepText string
	sepTerm string
	label   string
	body    []Node
}

func pruneBlock(b *Block) []Node {
	comps := []compartment{{first: true, label: b.Label, body: pruneNodes(b.Body)}}
	for _, e := range b.Elses {
		comps = append(comps, compartment{
			sepText: e.SepText,
			sepTerm: e.SepTerm,
			label:   e.Label,
			body:    pruneNodes(e.Body),
		})
	}

	var survivors []compartment
	for _, c := range comps {
		if hasContent(c.body) {
			survivors = append(survivors, c)
		}
	}

	if len(survivors) == 0 {
		return nil
	}

	if len(survivors) == 1 && (b.Keyword == "par" || b.Keyword == "alt") {
		body := survivors[0].body
		return dedentNodes(body, unwrapPrefix(b, body))
	}

	nb := *b
	first := survivors[0]
	if first.first {
		nb.HeaderText, nb.HeaderTerm, nb.Label = b.HeaderText, b.HeaderTerm, b.Label
	} else {
		nb.HeaderText, nb.HeaderTerm, nb.Label = promotedHeader(b, first.label), b.HeaderTerm, first.label
	}
	nb.Body = first.body
	nb.Elses = nil
	for _, c := range survivors[1:] {
		nb.Elses = append(nb.Elses, ElseClause{SepText: c.sepText, SepTerm: c.sepTerm, Label: c.label, Body: c.body})
	}
	return []Node{&nb}
}

func hasContent(nodes []Node) bool {
	for _, n := range nodes {
		switch n.(type) {
		case *Message, *Block, *Ref:
			return true
		}
	}
	return false
}

func promotedHeader(b *Block, label string) string {
	head := lines.LeadingIndent(b.HeaderText) + b.Keyword
	if label != "" {
		head += " " + label
	}
	return head
}

func unwrapPrefix(b *Block, body []Node) string {
	headerIndent := lines.LeadingIndent(b.HeaderText)
	bodyIndent, ok := minBodyIndent(body)
	if !ok || !strings.HasPrefix(bodyIndent, headerIndent) {
		return ""
	}
	return strings.TrimPrefix(bodyIndent, headerIndent)
}

func minBodyIndent(body []Node) (string, bool) {
	min := ""
	found := false
	for _, n := range body {
		ind, ok := nodeIndent(n)
		if !ok {
			continue
		}
		if !found || len(ind) < len(min) {
			min, found = ind, true
		}
	}
	return min, found
}

func nodeIndent(n Node) (string, bool) {
	switch v := n.(type) {
	case *Message:
		return v.Indent, true
	case *Block:
		return lines.LeadingIndent(v.HeaderText), true
	case *Note:
		return lines.LeadingIndent(v.HeaderPrefix), true
	case *Ref:
		return lines.LeadingIndent(v.HeaderPrefix), true
	case *Delay:
		return lines.LeadingIndent(v.Text), true
	case *Activation:
		return lines.LeadingIndent(v.Before), true
	default:
		return "", false
	}
}

func dedentNodes(nodes []Node, prefix string) []Node {
	if prefix == "" {
		return nodes
	}
	out := make([]Node, len(nodes))
	for i, n := range nodes {
		out[i] = dedentNode(n, prefix)
	}
	return out
}

func dedentNode(n Node, prefix string) Node {
	switch v := n.(type) {
	case *Message:
		nm := *v
		nm.Indent = strings.TrimPrefix(nm.Indent, prefix)
		return &nm
	case *Note:
		nn := *v
		nn.HeaderPrefix = strings.TrimPrefix(nn.HeaderPrefix, prefix)
		nn.Body = stripLinePrefix(nn.Body, prefix)
		nn.EndText = strings.TrimPrefix(nn.EndText, prefix)
		return &nn
	case *Ref:
		nr := *v
		nr.HeaderPrefix = strings.TrimPrefix(nr.HeaderPrefix, prefix)
		nr.Body = stripLinePrefix(nr.Body, prefix)
		nr.EndText = strings.TrimPrefix(nr.EndText, prefix)
		return &nr
	case *Delay:
		nd := *v
		nd.Text = strings.TrimPrefix(nd.Text, prefix)
		return &nd
	case *Activation:
		na := *v
		na.Before = strings.TrimPrefix(na.Before, prefix)
		return &na
	case *Block:
		nb := *v
		nb.HeaderText = strings.TrimPrefix(nb.HeaderText, prefix)
		nb.EndText = strings.TrimPrefix(nb.EndText, prefix)
		nb.Body = dedentNodes(v.Body, prefix)
		if v.Elses != nil {
			nb.Elses = make([]ElseClause, len(v.Elses))
			for i, e := range v.Elses {
				ne := e
				ne.SepText = strings.TrimPrefix(ne.SepText, prefix)
				ne.Body = dedentNodes(e.Body, prefix)
				nb.Elses[i] = ne
			}
		}
		return &nb
	case *Raw:
		nr := *v
		nr.Text = stripLinePrefix(nr.Text, prefix)
		return &nr
	default:
		return n
	}
}

func stripLinePrefix(s, prefix string) string {
	if s == "" || prefix == "" {
		return s
	}
	var b strings.Builder
	for len(s) > 0 {
		line, term := s, ""
		if i := strings.IndexByte(s, '\n'); i >= 0 {
			line, term, s = s[:i], "\n", s[i+1:]
		} else {
			s = ""
		}
		b.WriteString(strings.TrimPrefix(line, prefix))
		b.WriteString(term)
	}
	return b.String()
}
