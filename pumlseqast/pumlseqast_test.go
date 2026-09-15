package pumlseqast

import (
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseRoundTrip(t *testing.T) {
	tests := map[string]string{
		"empty":                        "",
		"header only":                  "@startuml\n@enduml\n",
		"no trailing newline":          "@startuml\n@enduml",
		"blank line between":           "@startuml\n\nparticipant A\n\n@enduml\n",
		"trailing blank lines":         "@startuml\n@enduml\n\n\n",
		"comment after enduml":         "@startuml\n@enduml\n' trailing comment\n",
		"crlf":                         "@startuml\r\nA -> B\r\n@enduml\r\n",
		"named startuml":               "@startuml Seq\nA -> B\n@enduml\n",
		"comments and directives":      "@startuml\n' a comment\nautonumber\ntitle My Title\nskinparam monochrome true\n@enduml\n",
		"line comment with arrow text": "@startuml\na -> b\n'a -> b\nb -> c\n@enduml\n",
		"single line block comment":    "@startuml\n/' commented '/\nA -> B\n@enduml\n",
		"multiline block comment":      "@startuml\na -> b\n/'\nx -> y\n'/\nb -> c\n@enduml\n",
		"delay and divider":            "@startuml\na -> b\n...\nc -> d\n== Phase 2 ==\ne -> f\n...5 minutes later...\ng -> h\n@enduml\n",
		"participants":                 "@startuml\nparticipant A\nactor User\ndatabase \"DB\" as db\nparticipant \"Long Name\" as LN\n  participant Indented\n@enduml\n",
		"messages":                     "@startuml\na -> b\na->b\nb <- a\na -->> b: hi\na ->x b\n  c -> d: indented\na -> a: self\n@enduml\n",
		"note single over":             "@startuml\nnote over A: hello\n@enduml\n",
		"note single left of":          "@startuml\nnote left of A: x\n@enduml\n",
		"note free text single":        "@startuml\nnote left: free text\n@enduml\n",
		"note multiline with arrow":    "@startuml\nnote over A, B\n  body line\n  a -> a looks like a message\nend note\n@enduml\n",
		"note multiline unterminated":  "@startuml\nnote over A\nbody\n@enduml\n",
		"ref single over":              "@startuml\nref over A, B : ./sub.puml\n@enduml\n",
		"ref multiline":                "@startuml\nref over A, B\n  ./sub.puml\nend ref\n@enduml\n",
		"ref multiline with arrow":     "@startuml\nref over A, B\n  a -> a looks like a message\nend ref\n@enduml\n",
		"ref multiline with else":      "@startuml\nalt x\n  ref over A, B\n    else looks like a clause\n  end ref\nend\n@enduml\n",
		"ref multiline unterminated":   "@startuml\nref over A\nbody\n@enduml\n",
		"ref endref spelling":          "@startuml\nref over A\n  ./sub.puml\nendref\n@enduml\n",
		"ref crlf":                     "@startuml\r\nref over A, B\r\n  ./sub.puml\r\nend ref\r\n@enduml\r\n",
		"ref single crlf":              "@startuml\r\nref over A, B : ./sub.puml\r\n@enduml\r\n",
		"ref indented":                 "@startuml\ngroup G\n  ref over A, B : ./sub.puml\nend\n@enduml\n",
		"activations":                  "@startuml\nactivate A\ndeactivate A\ndestroy B\ncreate C\nA -> C\n@enduml\n",
		"activation forms":             "@startuml\nactivate A\nactivate   Z\ncreate \"Long Name\"\ncreate participant \"Svc X\" as Y\n@enduml\n",
		"group":                        "@startuml\ngroup Auth\n  A -> B\nend\n@enduml\n",
		"nested group":                 "@startuml\ngroup Outer\n  A -> B\n  group Inner\n    B -> C\n  end\nend\n@enduml\n",
		"alt else":                     "@startuml\nalt success\n  A -> B\nelse failure\n  A -> C\nend\n@enduml\n",
		"multiple else":                "@startuml\nalt a\n  A -> B\nelse b\n  A -> C\nelse c\n  A -> D\nend\n@enduml\n",
		"par else":                     "@startuml\npar\n  A -> B\nelse\n  C -> D\nend\n@enduml\n",
		"nested alt inside else":       "@startuml\nalt outer\n  alt inner\n    A -> B\n  else inner2\n    A -> C\n  end\nelse outer2\n  D -> E\nend\n@enduml\n",
		"unbalanced group":             "@startuml\ngroup NoEnd\n  A -> B\n@enduml\n",
		"box degrades to raw":          "@startuml\nbox \"Group\"\n  participant A\nend box\nA -> B\n@enduml\n",
		"stray else at top level":      "@startuml\nelse stray\nA -> B\n@enduml\n",
		"fixture auth":                 "@startuml\nparticipant A\nparticipant B\nparticipant C\nA -> B: login\ngroup Auth\n  B -> C: verify\n  C -> B: ok\nend\nB -> A: done\n@enduml\n",
		"fixture nested":               "@startuml\nparticipant A\nparticipant B\nparticipant C\ngroup Outer\n  A -> B\n  group Inner\n    B -> C\n  end\nend\n@enduml\n",
	}

	for name, src := range tests {
		t.Run(name, func(t *testing.T) {
			d, err := Parse(src)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := d.String(); got != src {
				t.Errorf("round trip mismatch (-want +got):\n%s", cmp.Diff(src, got))
			}
		})
	}
}

func TestParseStructure(t *testing.T) {
	t.Run("message direction and self", func(t *testing.T) {
		d, _ := Parse("@startuml\nb <- a\nx -> x: self\n@enduml\n")
		msgs := messagesOf(d.Nodes)
		if len(msgs) != 2 {
			t.Fatalf("want 2 messages, got %d", len(msgs))
		}
		if msgs[0].Sender != "a" || msgs[0].Receiver != "b" {
			t.Errorf("reverse arrow: got sender=%q receiver=%q", msgs[0].Sender, msgs[0].Receiver)
		}
		if msgs[1].Sender != "x" || msgs[1].Receiver != "x" {
			t.Errorf("self message: got sender=%q receiver=%q", msgs[1].Sender, msgs[1].Receiver)
		}
	})

	t.Run("block then and else clauses separated", func(t *testing.T) {
		d, _ := Parse("@startuml\nalt a\n  A -> B\nelse b\n  A -> C\nelse c\n  A -> D\nend\n@enduml\n")
		var blk *Block
		for _, n := range d.Nodes {
			if b, ok := n.(*Block); ok {
				blk = b
			}
		}
		if blk == nil {
			t.Fatal("no block found")
		}
		if blk.Keyword != "alt" || blk.Label != "a" {
			t.Errorf("header: keyword=%q label=%q", blk.Keyword, blk.Label)
		}
		if len(blk.Body) != 1 {
			t.Errorf("then body: want 1 node, got %d", len(blk.Body))
		}
		if len(blk.Elses) != 2 {
			t.Fatalf("else clauses: want 2, got %d", len(blk.Elses))
		}
		if blk.Elses[0].Label != "b" || blk.Elses[1].Label != "c" {
			t.Errorf("else labels: %q, %q", blk.Elses[0].Label, blk.Elses[1].Label)
		}
	})

	t.Run("nested alt else does not split outer", func(t *testing.T) {
		d, _ := Parse("@startuml\nalt outer\n  alt inner\n    A -> B\n  else inner2\n    A -> C\n  end\nelse outer2\n  D -> E\nend\n@enduml\n")
		var outer *Block
		for _, n := range d.Nodes {
			if b, ok := n.(*Block); ok {
				outer = b
			}
		}
		if outer == nil {
			t.Fatal("no outer block")
		}
		if len(outer.Elses) != 1 {
			t.Fatalf("outer else clauses: want 1, got %d", len(outer.Elses))
		}
		if outer.Elses[0].Label != "outer2" {
			t.Errorf("outer else label: %q", outer.Elses[0].Label)
		}
	})

	t.Run("delay is a Delay node and divider stays Raw", func(t *testing.T) {
		d, _ := Parse("@startuml\na -> b\n...5 minutes later...\n== Phase 2 ==\nc -> d\n@enduml\n")
		var delays, raws int
		for _, n := range d.Nodes {
			switch n.(type) {
			case *Delay:
				delays++
			case *Raw:
				raws++
			}
		}
		if delays != 1 {
			t.Errorf("want 1 Delay node, got %d", delays)
		}
		if raws != 1 {
			t.Errorf("want 1 Raw node (the divider), got %d", raws)
		}
	})

	t.Run("activation target resolved across forms", func(t *testing.T) {
		cases := map[string]string{
			"@startuml\nactivate A\n@enduml\n":                        "A",
			"@startuml\ncreate \"Long Name\"\n@enduml\n":              "Long Name",
			"@startuml\ncreate participant Bob\n@enduml\n":            "Bob",
			"@startuml\ncreate participant \"Svc X\" as Y\n@enduml\n": "Y",
		}
		for src, want := range cases {
			d, _ := Parse(src)
			var act *Activation
			for _, n := range d.Nodes {
				if a, ok := n.(*Activation); ok {
					act = a
				}
			}
			if act == nil {
				t.Fatalf("%q: no Activation node", src)
			}
			if act.Target != want {
				t.Errorf("%q: Target=%q want %q", src, act.Target, want)
			}
		}
	})

	t.Run("note targets parsed", func(t *testing.T) {
		d, _ := Parse("@startuml\nnote over A, B: hi\n@enduml\n")
		var note *Note
		for _, n := range d.Nodes {
			if nn, ok := n.(*Note); ok {
				note = nn
			}
		}
		if note == nil {
			t.Fatal("no note found")
		}
		if !slices.Equal([]string{"A", "B"}, note.Targets) {
			t.Errorf("targets mismatch (-want +got):\n%s", cmp.Diff([]string{"A", "B"}, note.Targets))
		}
		if !note.Single {
			t.Error("want single-line note")
		}
	})

	t.Run("a colon inside quotes is not treated as a single-line note/ref marker", func(t *testing.T) {
		cases := map[string]string{
			"@startuml\nnote over \"A:1\"\n  hello\nend note\n@enduml\n": "  hello\n",
			"@startuml\nref over \"A:1\"\n  hello\nend ref\n@enduml\n":   "  hello\n",
		}
		for src, wantBody := range cases {
			d, _ := Parse(src)
			var single bool
			var body string
			for _, n := range d.Nodes {
				switch v := n.(type) {
				case *Note:
					single, body = v.Single, v.Body
				case *Ref:
					single, body = v.Single, v.Body
				}
			}
			if single {
				t.Errorf("%q: want multiline", src)
			}
			if body != wantBody {
				t.Errorf("%q: Body=%q want %q", src, body, wantBody)
			}
		}
	})

	t.Run("targets raw excludes trailing space and CR so renaming keeps them", func(t *testing.T) {
		cases := map[string]struct {
			Src    string
			Raw    string
			Suffix string
		}{
			"space just before the colon is folded into the suffix": {Src: "@startuml\nnote over A, B : hi\n@enduml\n", Raw: "A, B", Suffix: " : hi"},
			"CR of CRLF is folded into the suffix":                  {Src: "@startuml\r\nnote over A, B\r\nx\r\nend note\r\n@enduml\r\n", Raw: "A, B", Suffix: "\r"},
		}
		for name, want := range cases {
			t.Run(name, func(t *testing.T) {
				d, _ := Parse(want.Src)
				var note *Note
				for _, n := range d.Nodes {
					if nn, ok := n.(*Note); ok {
						note = nn
					}
				}
				if note == nil {
					t.Fatal("no note found")
				}
				if note.TargetsRaw != want.Raw || note.HeaderSuffix != want.Suffix {
					t.Errorf("TargetsRaw=%q HeaderSuffix=%q want %q / %q", note.TargetsRaw, note.HeaderSuffix, want.Raw, want.Suffix)
				}
			})
		}
	})

	t.Run("participant label derived from the declaration", func(t *testing.T) {
		cases := map[string]struct {
			Src      string
			Expected string
		}{
			"no alias keeps the declared name":    {Src: "@startuml\nparticipant A\n@enduml\n", Expected: "A"},
			"quoted display name with an alias":   {Src: "@startuml\nparticipant \"Cache\" as C\n@enduml\n", Expected: "Cache"},
			"unquoted display name with an alias": {Src: "@startuml\nparticipant A as B\n@enduml\n", Expected: "A"},
			"actor with a quoted display name":    {Src: "@startuml\nactor \"The User\" as u\n@enduml\n", Expected: "The User"},
			"database keyword":                    {Src: "@startuml\ndatabase DB\n@enduml\n", Expected: "DB"},
		}
		for name, want := range cases {
			t.Run(name, func(t *testing.T) {
				d, _ := Parse(want.Src)
				var p *Participant
				for _, n := range d.Nodes {
					if pp, ok := n.(*Participant); ok {
						p = pp
					}
				}
				if p == nil {
					t.Fatal("no participant found")
				}
				if got := p.Label(); got != want.Expected {
					t.Errorf("Label()=%q want %q", got, want.Expected)
				}
			})
		}
	})

	t.Run("ref targets parsed and single line label", func(t *testing.T) {
		d, _ := Parse("@startuml\nref over A, B : ./sub.puml\n@enduml\n")
		ref := soleRef(t, d.Nodes)
		if !slices.Equal([]string{"A", "B"}, ref.Targets) {
			t.Errorf("targets mismatch (-want +got):\n%s", cmp.Diff([]string{"A", "B"}, ref.Targets))
		}
		if !ref.Single {
			t.Error("want single-line ref")
		}
		if got := ref.Label(); got != "./sub.puml" {
			t.Errorf("Label()=%q want %q", got, "./sub.puml")
		}
		if ref.Line != 2 {
			t.Errorf("Line=%d want 2", ref.Line)
		}
	})

	t.Run("ref targets raw excludes trailing space and CR so renaming keeps them", func(t *testing.T) {
		cases := map[string]struct {
			Src    string
			Raw    string
			Suffix string
		}{
			"space just before the colon is folded into the suffix": {Src: "@startuml\nref over A, B : ./sub.puml\n@enduml\n", Raw: "A, B", Suffix: " : ./sub.puml"},
			"CR of CRLF is folded into the suffix":                  {Src: "@startuml\r\nref over A, B\r\n./sub.puml\r\nend ref\r\n@enduml\r\n", Raw: "A, B", Suffix: "\r"},
		}
		for name, want := range cases {
			t.Run(name, func(t *testing.T) {
				d, _ := Parse(want.Src)
				ref := soleRef(t, d.Nodes)
				if ref.TargetsRaw != want.Raw || ref.HeaderSuffix != want.Suffix {
					t.Errorf("TargetsRaw=%q HeaderSuffix=%q want %q / %q", ref.TargetsRaw, ref.HeaderSuffix, want.Raw, want.Suffix)
				}
			})
		}
	})

	t.Run("ref body is opaque so its arrows are not messages", func(t *testing.T) {
		d, _ := Parse("@startuml\nref over A, B\n  a -> a looks like a message\nend ref\n@enduml\n")
		if msgs := messagesOf(d.Nodes); len(msgs) != 0 {
			t.Errorf("want 0 messages, got %d", len(msgs))
		}
		ref := soleRef(t, d.Nodes)
		if ref.Single {
			t.Error("want multiline ref")
		}
		if got := ref.Label(); got != "a -> a looks like a message" {
			t.Errorf("Label()=%q", got)
		}
	})

	t.Run("ref body else does not split the enclosing block", func(t *testing.T) {
		d, _ := Parse("@startuml\nalt x\n  ref over A, B\n    else looks like a clause\n  end ref\nend\n@enduml\n")
		var block *Block
		for _, n := range d.Nodes {
			if b, ok := n.(*Block); ok {
				block = b
			}
		}
		if block == nil {
			t.Fatal("no block found")
		}
		if len(block.Elses) != 0 {
			t.Errorf("want 0 else clauses, got %d", len(block.Elses))
		}
		soleRef(t, block.Body)
	})

	t.Run("multiline ref label joins body lines", func(t *testing.T) {
		d, _ := Parse("@startuml\nref over A\n  first\n\n  second\nend ref\n@enduml\n")
		ref := soleRef(t, d.Nodes)
		if got := ref.Label(); got != "first\nsecond" {
			t.Errorf("Label()=%q want %q", got, "first\nsecond")
		}
	})
}

func soleRef(t *testing.T, nodes []Node) *Ref {
	t.Helper()
	refs := RefsIn(nodes)
	if len(refs) != 1 {
		t.Fatalf("want 1 Ref node, got %d", len(refs))
	}
	return refs[0]
}

func messagesOf(nodes []Node) []*Message {
	var out []*Message
	for _, n := range nodes {
		if m, ok := n.(*Message); ok {
			out = append(out, m)
		}
	}
	return out
}
