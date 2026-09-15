package csp

import (
	"slices"
	"testing"

	"github.com/Kuniwak/puml-tools/pumlseqast"
)

var (
	x = pumlseqast.MessageKey{Sender: "A", Receiver: "B", Label: "x"}
	y = pumlseqast.MessageKey{Sender: "A", Receiver: "B", Label: "y"}
)

func TestEpsClosure(t *testing.T) {

	build := func() *LTS {
		l := NewLTS()
		s1, s2, s3 := l.AddState(), l.AddState(), l.AddState()
		l.AddEps(0, s1)
		l.AddEps(s1, s2)
		l.AddEps(s2, 0)
		l.AddMove(s1, x, s3)
		return l
	}

	tests := map[string]struct {
		States   []int
		Expected []int
	}{
		"cycle is closed without repeating":   {States: []int{0}, Expected: []int{0, 1, 2}},
		"isolated state closes to itself":     {States: []int{3}, Expected: []int{3}},
		"labeled moves are not followed":      {States: []int{1}, Expected: []int{0, 1, 2}},
		"union of the given states is sorted": {States: []int{3, 0}, Expected: []int{0, 1, 2, 3}},
		"duplicated input is deduplicated":    {States: []int{0, 0}, Expected: []int{0, 1, 2}},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := build().EpsClosure(test.States); !slices.Equal(test.Expected, got) {
				t.Errorf("EpsClosure(%v) = %v, want %v", test.States, got, test.Expected)
			}
		})
	}
}

func TestDeterminizeTraces(t *testing.T) {

	l := NewLTS()
	b1, b2 := l.AddState(), l.AddState()
	l.AddEps(0, b1)
	l.AddEps(0, b2)
	e1, e2, e3 := l.AddState(), l.AddState(), l.AddState()
	l.AddMove(b1, x, e1)
	l.AddMove(b2, x, e2)
	l.AddMove(e2, y, e3)
	d := l.DeterminizeTraces()

	tests := map[string]struct {
		Trace    []pumlseqast.MessageKey
		Expected []pumlseqast.MessageKey
	}{
		"start allows only x":        {Trace: nil, Expected: []pumlseqast.MessageKey{x}},
		"after x both branches live": {Trace: []pumlseqast.MessageKey{x}, Expected: []pumlseqast.MessageKey{y}},
		"after x y nothing is left":  {Trace: []pumlseqast.MessageKey{x, y}, Expected: []pumlseqast.MessageKey{}},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			state := d.Start()
			for _, label := range test.Trace {
				next, ok := d.Next(state, label)
				if !ok {
					t.Fatalf("trace %v is rejected at %v", test.Trace, label)
				}
				state = next
			}
			if got := d.Labels(state); !slices.Equal(test.Expected, got) {
				t.Errorf("Labels after %v = %v, want %v", test.Trace, got, test.Expected)
			}
		})
	}
}

func TestCompileSource(t *testing.T) {
	tests := map[string]struct {
		Src         string
		Expected    []pumlseqast.MessageKey
		ExpectedErr bool
	}{
		"single message": {
			Src:      "@startuml\nA -> B: x\n@enduml\n",
			Expected: []pumlseqast.MessageKey{x},
		},
		"choice offers both branches": {
			Src:      "@startuml\nalt ok\nA -> B: x\nelse ng\nA -> B: y\nend\n@enduml\n",
			Expected: []pumlseqast.MessageKey{x, y},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			l, err := CompileSource(test.Src)
			if (err != nil) != test.ExpectedErr {
				t.Fatalf("CompileSource returned error %v, want error = %v", err, test.ExpectedErr)
			}
			d := l.DeterminizeTraces()
			if got := d.Labels(d.Start()); !slices.Equal(test.Expected, got) {
				t.Errorf("labels at start = %v, want %v", got, test.Expected)
			}
		})
	}
}

func TestLTSRejectsUnknownState(t *testing.T) {
	tests := map[string]func(l *LTS){
		"AddEps from":  func(l *LTS) { l.AddEps(9, 0) },
		"AddEps to":    func(l *LTS) { l.AddEps(0, 9) },
		"AddMove from": func(l *LTS) { l.AddMove(9, x, 0) },
		"AddMove to":   func(l *LTS) { l.AddMove(0, x, 9) },
		"SetStart":     func(l *LTS) { l.SetStart(9) },
		"Eps":          func(l *LTS) { l.Eps(9) },
		"Moves":        func(l *LTS) { l.Moves(9) },
		"EpsClosure":   func(l *LTS) { l.EpsClosure([]int{9}) },
		"negative id":  func(l *LTS) { l.AddEps(-1, 0) },
	}

	for name, call := range tests {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("expected a panic for an unknown state")
				}
			}()
			call(NewLTS())
		})
	}
}

func TestCompilePanicsOnRef(t *testing.T) {
	tests := map[string]string{
		"single-line ref":    "@startuml\nA -> B: x\nref over A, B : ./sub.puml\n@enduml\n",
		"multi-line ref":     "@startuml\nref over A, B\n  ./sub.puml\nend ref\n@enduml\n",
		"ref inside a block": "@startuml\nalt c\n  ref over A, B : ./sub.puml\nend\n@enduml\n",
	}

	for name, src := range tests {
		t.Run(name, func(t *testing.T) {
			d, err := pumlseqast.Parse(src)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer func() {
				if recover() == nil {
					t.Error("Compile should panic on an unexpanded ref, but returned")
				}
			}()
			Compile(d)
		})
	}
}
