package pumlseqbreak

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Kuniwak/puml-tools/pumlseqast"
)

type Kind string

const (
	KindMessage Kind = "message"
	KindRef     Kind = "ref"
)

type Break struct {
	Line      int
	Kind      Kind
	Lifelines []string
}

func MessageBreak(line int, sender string) Break {
	return Break{Line: line, Kind: KindMessage, Lifelines: []string{sender}}
}

func RefBreak(line int, targets []string) Break {
	return Break{Line: line, Kind: KindRef, Lifelines: targets}
}

func (b Break) Reason() string {
	return fmt.Sprintf("%s has no preceding trigger", b.subject())
}

func (b Break) subject() string {
	switch b.Kind {
	case KindMessage:
		return "sender " + strings.Join(b.Lifelines, ", ")
	case KindRef:
		return "ref over " + strings.Join(b.Lifelines, ", ")
	default:
		panic(fmt.Sprintf("pumlseqbreak: invalid kind %q", b.Kind))
	}
}

type context struct {
	lifelines [][]string
	reset     bool
}

func resetContext() context {
	return context{reset: true}
}

func event(lifelines ...string) context {
	return context{lifelines: [][]string{lifelines}}
}

func mergeContexts(cs []context) context {
	var out context
	for _, c := range cs {
		if c.reset {
			return resetContext()
		}
		out.lifelines = append(out.lifelines, c.lifelines...)
	}
	return out
}

func (c context) connected(sender string) bool {
	if c.reset {
		return true
	}
	for _, ls := range c.lifelines {
		if slices.Contains(ls, sender) {
			return true
		}
	}
	return false
}

func Find(src []byte) []Break {
	d, err := pumlseqast.Parse(string(src))
	if err != nil {
		return nil
	}
	var breaks []Break
	walk(d.Nodes, resetContext(), &breaks)
	return breaks
}

func connectedAny(ctx context, lifelines []string) bool {
	for _, l := range lifelines {
		if ctx.connected(l) {
			return true
		}
	}
	return false
}

func walk(nodes []pumlseqast.Node, ctx context, breaks *[]Break) context {
	for _, n := range nodes {
		switch v := n.(type) {
		case *pumlseqast.Message:
			if !ctx.connected(v.Sender) {
				*breaks = append(*breaks, MessageBreak(v.Line, v.Sender))
			}
			ctx = event(v.Sender, v.Receiver)
		case *pumlseqast.Block:
			ctx = walkBlock(v, ctx, breaks)
		case *pumlseqast.Ref:

			if len(v.Targets) == 0 {
				continue
			}
			if !connectedAny(ctx, v.Targets) {
				*breaks = append(*breaks, RefBreak(v.Line, v.Targets))
			}
			ctx = event(v.Targets...)
		case *pumlseqast.Delay:
			ctx = resetContext()
		}
	}
	return ctx
}

func walkBlock(b *pumlseqast.Block, entry context, breaks *[]Break) context {
	clauseEnds := []context{walk(b.Body, entry, breaks)}
	for _, e := range b.Elses {
		clauseEnds = append(clauseEnds, walk(e.Body, entry, breaks))
	}
	if includesEntry(b.Keyword, len(b.Elses) > 0) {
		clauseEnds = append(clauseEnds, entry)
	}
	return mergeContexts(clauseEnds)
}

func includesEntry(keyword string, sawElse bool) bool {
	switch keyword {
	case "opt", "loop", "break":
		return true
	case "alt":
		return !sawElse
	default:
		return false
	}
}
