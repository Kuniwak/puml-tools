package csp

import (
	"slices"
	"strconv"
	"strings"

	"github.com/Kuniwak/puml-tools/ints"
	"github.com/Kuniwak/puml-tools/pumlseqast"
)

type Move struct {
	Label pumlseqast.MessageKey
	To    int
}

type LTS struct {
	start int
	eps   [][]int
	moves [][]Move
}

type frag struct {
	start  int
	accept int
}

func NewLTS() *LTS {
	l := &LTS{}
	l.AddState()
	return l
}

func (l *LTS) AddState() int {
	l.eps = append(l.eps, nil)
	l.moves = append(l.moves, nil)
	return len(l.eps) - 1
}

func (l *LTS) NumStates() int { return len(l.eps) }

func (l *LTS) Start() int { return l.start }

func (l *LTS) SetStart(s int) {
	l.mustState(s)
	l.start = s
}

func (l *LTS) AddEps(from, to int) {
	l.mustState(from)
	l.mustState(to)
	l.eps[from] = append(l.eps[from], to)
}

func (l *LTS) AddMove(from int, label pumlseqast.MessageKey, to int) {
	l.mustState(from)
	l.mustState(to)
	l.moves[from] = append(l.moves[from], Move{Label: label, To: to})
}

func (l *LTS) Eps(state int) []int {
	l.mustState(state)
	return slices.Clone(l.eps[state])
}

func (l *LTS) Moves(state int) []Move {
	l.mustState(state)
	return slices.Clone(l.moves[state])
}

func (l *LTS) mustState(s int) {
	if s < 0 || s >= len(l.eps) {
		panic("csp: no such state: " + strconv.Itoa(s))
	}
}

func CompileSource(src string) (*LTS, error) {
	d, err := pumlseqast.Parse(src)
	if err != nil {
		return nil, err
	}
	return Compile(d), nil
}

func Compile(d *pumlseqast.Diagram) *LTS {
	l := &LTS{}
	f := l.compileSeq(d.Nodes)
	l.start = f.start
	return l
}

func (l *LTS) compileSeq(nodes []pumlseqast.Node) frag {
	frags := make([]frag, len(nodes))
	for i, n := range nodes {
		frags[i] = l.compileNode(n)
	}
	return l.compileProduct(frags, l.weakGuard())
}

func (l *LTS) weakGuard() guard {
	emitters := l.emitters()
	canEmit := func(participant string, state int) bool {
		states, ok := emitters[participant]
		return ok && states[state]
	}
	return func(t []int, i int, k pumlseqast.MessageKey) bool {
		for j := 0; j < i; j++ {
			if canEmit(k.Sender, t[j]) || canEmit(k.Receiver, t[j]) {
				return false
			}
		}
		return true
	}
}

func (l *LTS) emitters() map[string][]bool {
	n := len(l.eps)
	rev := make([][]int, n)
	seeds := map[string][]int{}
	for s := 0; s < n; s++ {
		for _, to := range l.eps[s] {
			rev[to] = append(rev[to], s)
		}
		for _, mv := range l.moves[s] {
			rev[mv.To] = append(rev[mv.To], s)
			seeds[mv.Label.Sender] = append(seeds[mv.Label.Sender], s)
			seeds[mv.Label.Receiver] = append(seeds[mv.Label.Receiver], s)
		}
	}

	out := make(map[string][]bool, len(seeds))
	for participant, queue := range seeds {
		states := make([]bool, n)
		for _, s := range queue {
			states[s] = true
		}
		for len(queue) > 0 {
			s := queue[0]
			queue = queue[1:]
			for _, from := range rev[s] {
				if states[from] {
					continue
				}
				states[from] = true
				queue = append(queue, from)
			}
		}
		out[participant] = states
	}
	return out
}

type guard func(t []int, i int, k pumlseqast.MessageKey) bool

func (l *LTS) compileProduct(frags []frag, g guard) frag {
	if len(frags) == 0 {
		s := l.AddState()
		return frag{start: s, accept: s}
	}

	startTuple := make([]int, len(frags))
	acceptTuple := make([]int, len(frags))
	for i, f := range frags {
		startTuple[i] = f.start
		acceptTuple[i] = f.accept
	}

	memo := map[string]int{}
	var queue [][]int
	getOrCreate := func(t []int) int {
		key := ints.Key(t)
		if id, ok := memo[key]; ok {
			return id
		}
		id := l.AddState()
		memo[key] = id
		queue = append(queue, append([]int(nil), t...))
		return id
	}

	startID := getOrCreate(startTuple)
	acceptID := getOrCreate(acceptTuple)
	for len(queue) > 0 {
		t := queue[0]
		queue = queue[1:]
		id := memo[ints.Key(t)]
		for i, si := range t {
			for _, to := range l.eps[si] {
				l.AddEps(id, getOrCreate(ints.Replaced(t, i, to)))
			}
			for _, mv := range l.moves[si] {
				if g != nil && !g(t, i, mv.Label) {
					continue
				}
				l.AddMove(id, mv.Label, getOrCreate(ints.Replaced(t, i, mv.To)))
			}
		}
	}
	return frag{start: startID, accept: acceptID}
}

func (l *LTS) compileNode(n pumlseqast.Node) frag {
	switch v := n.(type) {
	case *pumlseqast.Message:
		s0 := l.AddState()
		s1 := l.AddState()
		l.AddMove(s0, v.Key(), s1)
		return frag{start: s0, accept: s1}
	case *pumlseqast.Block:
		return l.compileBlock(v)
	case *pumlseqast.Ref:
		panic("csp: unexpanded ref over " + strings.Join(v.Targets, ", ") +
			" (call pumlseqast.ValidateNoRef first)")
	default:
		s := l.AddState()
		return frag{start: s, accept: s}
	}
}

func (l *LTS) compileBlock(b *pumlseqast.Block) frag {
	switch b.Keyword {
	case "alt":
		return l.compileChoice(branchBodies(b), false)
	case "opt":
		return l.compileChoice(branchBodies(b), true)
	case "loop":
		return l.compileLoop(b.Body)
	case "par":
		return l.compilePar(branchBodies(b))
	case "group", "critical", "break":

		return l.compileSeq(b.Body)
	default:

		panic("csp: unsupported block keyword: " + b.Keyword)
	}
}

func branchBodies(b *pumlseqast.Block) [][]pumlseqast.Node {
	bodies := [][]pumlseqast.Node{b.Body}
	for _, e := range b.Elses {
		bodies = append(bodies, e.Body)
	}
	return bodies
}

func (l *LTS) compileChoice(bodies [][]pumlseqast.Node, withSkip bool) frag {
	s := l.AddState()
	e := l.AddState()
	for _, body := range bodies {
		f := l.compileSeq(body)
		l.AddEps(s, f.start)
		l.AddEps(f.accept, e)
	}
	if withSkip {
		l.AddEps(s, e)
	}
	return frag{start: s, accept: e}
}

func (l *LTS) compileLoop(body []pumlseqast.Node) frag {
	s := l.AddState()
	f := l.compileSeq(body)
	l.AddEps(s, f.start)
	l.AddEps(f.accept, s)
	return frag{start: s, accept: s}
}

func (l *LTS) compilePar(bodies [][]pumlseqast.Node) frag {
	frags := make([]frag, len(bodies))
	for i, body := range bodies {
		frags[i] = l.compileSeq(body)
	}
	return l.compileProduct(frags, nil)
}
