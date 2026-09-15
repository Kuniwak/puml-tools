package pumlseqast

func SetName(d *Diagram, name string) *Diagram {
	out := make([]Node, len(d.Nodes))
	for i, n := range d.Nodes {
		if s, ok := n.(*Start); ok {
			out[i] = newStart(name, s.Term)
			continue
		}
		out[i] = n
	}
	return &Diagram{Nodes: out}
}
