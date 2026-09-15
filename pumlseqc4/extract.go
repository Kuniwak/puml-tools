package pumlseqc4

import (
	"fmt"
	"strings"

	"github.com/Kuniwak/puml-tools/pumlseqast"
)

const (
	KindPerson = "Person"
	KindSystem = "System"
)

type participant struct {
	ID       string
	Label    string
	Kind     string
	Declared bool
}

type message struct {
	Src   string
	Dst   string
	Label string
}

func extractSeq(d *pumlseqast.Diagram) ([]participant, []message) {
	var order []participant
	seen := make(map[string]bool)
	add := func(p participant) {
		if seen[p.ID] {
			return
		}
		seen[p.ID] = true
		order = append(order, p)
	}

	var walk func(nodes []pumlseqast.Node)
	walk = func(nodes []pumlseqast.Node) {
		for _, n := range nodes {
			switch v := n.(type) {
			case *pumlseqast.Participant:
				add(participant{ID: v.Name, Label: v.Label(), Kind: kindOf(v.Keyword), Declared: true})
			case *pumlseqast.Block:
				walk(v.Body)
				for _, e := range v.Elses {
					walk(e.Body)
				}
			}
		}
	}
	walk(d.Nodes)

	var msgs []message
	for _, m := range pumlseqast.Messages(d) {
		msgs = append(msgs, message{Src: m.Sender, Dst: m.Receiver, Label: normalizeLabel(m.Key().Label)})
		add(participant{ID: m.Sender, Label: m.Sender, Kind: KindSystem})
		add(participant{ID: m.Receiver, Label: m.Receiver, Kind: KindSystem})
	}
	return order, msgs
}

func kindOf(keyword string) string {
	if keyword == "actor" {
		return KindPerson
	}
	return KindSystem
}

func normalizeLabel(s string) string {
	s = strings.TrimSpace(s)
	for strings.HasPrefix(s, "<") {
		end := strings.IndexByte(s, '>')
		if end < 0 {
			break
		}
		s = strings.TrimSpace(s[end+1:])
	}
	return s
}

func NormalizeBoxLabel(s string) string {
	s = normalizeLabel(s)
	for strings.HasSuffix(s, ">") {
		start := strings.LastIndex(s, "</")
		if start < 0 {
			break
		}
		s = strings.TrimSpace(s[:start])
		s = normalizeLabel(s)
	}
	return s
}

type RefWarning struct {
	Line    int
	Targets []string
}

func (w RefWarning) Reason() string {
	return fmt.Sprintf("ref over %s does not contribute to C4 (pass the referenced diagram as well, or run it through pumlseqexpand)",
		strings.Join(w.Targets, ", "))
}

func RefWarnings(d *pumlseqast.Diagram) []RefWarning {
	var out []RefWarning
	for _, r := range pumlseqast.Refs(d) {
		out = append(out, RefWarning{Line: r.Line, Targets: r.Targets})
	}
	return out
}
