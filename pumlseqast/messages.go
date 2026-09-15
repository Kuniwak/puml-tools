package pumlseqast

import (
	"fmt"
	"strings"
)

type MessageKey struct {
	Sender   string
	Receiver string
	Label    string
}

func (m MessageKey) String() string {
	if m.Label == "" {
		return m.Sender + " -> " + m.Receiver
	}
	return m.Sender + " -> " + m.Receiver + ": " + m.Label
}

func CompareMessageKey(a, b MessageKey) int {
	if a.Sender != b.Sender {
		return strings.Compare(a.Sender, b.Sender)
	}
	if a.Receiver != b.Receiver {
		return strings.Compare(a.Receiver, b.Receiver)
	}
	return strings.Compare(a.Label, b.Label)
}

func (m *Message) Key() MessageKey {
	return MessageKey{Sender: m.Sender, Receiver: m.Receiver, Label: messageLabel(m)}
}

func Messages(d *Diagram) []*Message {
	return MessagesIn(d.Nodes)
}

func MessagesIn(nodes []Node) []*Message {
	var out []*Message
	collectMessages(nodes, &out)
	return out
}

func collectMessages(nodes []Node, out *[]*Message) {
	for _, n := range nodes {
		switch v := n.(type) {
		case *Message:
			*out = append(*out, v)
		case *Block:
			collectMessages(v.Body, out)
			for _, e := range v.Elses {
				collectMessages(e.Body, out)
			}
		}
	}
}

func MessagesOutsideGroups(d *Diagram) []*Message {
	var out []*Message
	collectMessagesOutsideGroups(d.Nodes, &out)
	return out
}

func collectMessagesOutsideGroups(nodes []Node, out *[]*Message) {
	for _, n := range nodes {
		switch v := n.(type) {
		case *Message:
			*out = append(*out, v)
		case *Block:
			if v.Keyword == "group" {
				continue
			}
			collectMessagesOutsideGroups(v.Body, out)
			for _, e := range v.Elses {
				collectMessagesOutsideGroups(e.Body, out)
			}
		}
	}
}

func messageLabel(m *Message) string {
	if i := strings.Index(m.Rest, ":"); i >= 0 {
		return strings.TrimSpace(m.Rest[i+1:])
	}
	return ""
}

func Refs(d *Diagram) []*Ref {
	return RefsIn(d.Nodes)
}

func RefsIn(nodes []Node) []*Ref {
	var out []*Ref
	collectRefs(nodes, &out)
	return out
}

func collectRefs(nodes []Node, out *[]*Ref) {
	for _, n := range nodes {
		switch v := n.(type) {
		case *Ref:
			*out = append(*out, v)
		case *Block:
			collectRefs(v.Body, out)
			for _, e := range v.Elses {
				collectRefs(e.Body, out)
			}
		}
	}
}

func ValidateNoRef(name string, d *Diagram) error {
	refs := Refs(d)
	if len(refs) == 0 {
		return nil
	}
	r := refs[0]
	return fmt.Errorf("%s:%d: unexpanded ref found: ref over %s; run it through pumlseqexpand first",
		name, r.Line, strings.Join(r.Targets, ", "))
}
