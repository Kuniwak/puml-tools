package pumlseqast

func Hide(d *Diagram, hide func(*Message) bool) *Diagram {
	return &Diagram{Nodes: hideNodes(d.Nodes, hide)}
}

func hideNodes(nodes []Node, hide func(*Message) bool) []Node {
	var out []Node
	var pendingHide *Message
	for _, n := range nodes {
		switch v := n.(type) {
		case *Message:
			if hide(v) {
				pendingHide = v
				continue
			}
			out = append(out, v)
			pendingHide = nil
		case *Note:
			if pendingHide != nil && accompanies(v, pendingHide) {
				continue
			}
			out = append(out, v)

		case *Ref:
			out = append(out, v)
		case *Raw:
			out = append(out, v)
		case *Block:
			nb := *v
			nb.Body = hideNodes(v.Body, hide)
			if v.Elses != nil {
				nb.Elses = make([]ElseClause, len(v.Elses))
				for i, e := range v.Elses {
					ne := e
					ne.Body = hideNodes(e.Body, hide)
					nb.Elses[i] = ne
				}
			}
			out = append(out, &nb)
			pendingHide = nil
		default:
			out = append(out, n)
			pendingHide = nil
		}
	}
	return out
}

func accompanies(note *Note, msg *Message) bool {
	if len(note.Targets) == 0 {
		return true
	}
	for _, t := range note.Targets {
		if t != msg.Sender && t != msg.Receiver {
			return false
		}
	}
	return true
}
