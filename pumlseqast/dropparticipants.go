package pumlseqast

func DropParticipants(d *Diagram, drop func(*Participant) bool) *Diagram {
	return &Diagram{Nodes: dropParticipantNodes(d.Nodes, drop)}
}

func dropParticipantNodes(nodes []Node, drop func(*Participant) bool) []Node {
	var out []Node
	for _, n := range nodes {
		switch v := n.(type) {
		case *Participant:
			if drop(v) {
				continue
			}
			out = append(out, v)
		case *Block:
			nb := *v
			nb.Body = dropParticipantNodes(v.Body, drop)
			if v.Elses != nil {
				nb.Elses = make([]ElseClause, len(v.Elses))
				for i, e := range v.Elses {
					ne := e
					ne.Body = dropParticipantNodes(e.Body, drop)
					nb.Elses[i] = ne
				}
			}
			out = append(out, &nb)
		default:
			out = append(out, n)
		}
	}
	return out
}
