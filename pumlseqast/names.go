package pumlseqast

import "github.com/Kuniwak/puml-tools/sets"

func DeclaredNames(nodes []Node) sets.Set {
	set := sets.New()
	collectDeclaredNames(nodes, set)
	return set
}

func ReferencedNames(nodes []Node) sets.Set {
	set := sets.New()
	collectReferencedNames(nodes, set)
	return set
}

func collectDeclaredNames(nodes []Node, set sets.Set) {
	for _, n := range nodes {
		switch v := n.(type) {
		case *Participant:
			set.Add(v.Name)
		case *Block:
			collectDeclaredNames(v.Body, set)
			for _, e := range v.Elses {
				collectDeclaredNames(e.Body, set)
			}
		}
	}
}

func collectReferencedNames(nodes []Node, set sets.Set) {
	for _, n := range nodes {
		switch v := n.(type) {
		case *Message:
			set.Add(v.Sender)
			set.Add(v.Receiver)
		case *Activation:
			if v.Target != "" {
				set.Add(v.Target)
			}
		case *Note:
			for _, t := range v.Targets {
				set.Add(t)
			}
		case *Ref:
			for _, t := range v.Targets {
				set.Add(t)
			}
		case *Block:
			collectReferencedNames(v.Body, set)
			for _, e := range v.Elses {
				collectReferencedNames(e.Body, set)
			}
		}
	}
}
