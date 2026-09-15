package pumlseqast

import (
	"errors"
	"fmt"

	"github.com/Kuniwak/puml-tools/sets"
)

func Project(d *Diagram, ids []string) (*Diagram, error) {
	if len(ids) == 0 {
		return nil, errors.New("project requires at least one participant id")
	}
	present := sets.New(Participants(d)...)
	for _, id := range ids {
		if !present.Contains(id) {
			return nil, fmt.Errorf("participant not found: %s", id)
		}
	}

	keep := sets.New(ids...)
	projected := Hide(d, func(m *Message) bool {
		return !keep.Contains(m.Sender) && !keep.Contains(m.Receiver)
	})
	projected = Prune(projected)

	declared := sets.New(ids...)
	declared.AddAll(ReferencedNames(projected.Nodes))
	projected = DropParticipants(projected, func(p *Participant) bool {
		return !declared.Contains(p.Name)
	})
	return PruneEmptyBoxes(projected), nil
}
