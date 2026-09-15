package pumlseq

import (
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNoteTargets(t *testing.T) {
	tests := map[string]struct {
		Line     string
		Expected []string
	}{
		"note over single":       {Line: "note over A", Expected: []string{"A"}},
		"note over with message": {Line: "note over A: hello", Expected: []string{"A"}},
		"note over multiple":     {Line: "note over A, B", Expected: []string{"A", "B"}},
		"note left of":           {Line: "note left of A", Expected: []string{"A"}},
		"note right of":          {Line: "note right of A", Expected: []string{"A"}},
		"quoted target":          {Line: `note over "Svc One"`, Expected: []string{"Svc One"}},
		"not a positioned note":  {Line: "note this is free text", Expected: nil},

		"quoted target containing a colon":            {Line: `note over "A:1"`, Expected: []string{"A:1"}},
		"quoted target containing a colon with label": {Line: `note over "A:1" : hello`, Expected: []string{"A:1"}},
		"quoted ref target containing a colon":        {Line: `ref over "A:1" : ./sub.puml`, Expected: []string{"A:1"}},
		"label may contain further colons":            {Line: "note over A: text: with colon", Expected: []string{"A"}},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := NoteTargets(test.Line)
			if !slices.Equal(test.Expected, got) {
				t.Errorf("NoteTargets mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}

func TestSplitParticipants(t *testing.T) {
	tests := map[string]struct {
		Input    string
		Expected []string
	}{
		"single":          {Input: "A", Expected: []string{"A"}},
		"comma separated": {Input: "A, B", Expected: []string{"A", "B"}},
		"trims spaces":    {Input: "  A ,  B ", Expected: []string{"A", "B"}},
		"unquotes":        {Input: `"Svc One", B`, Expected: []string{"Svc One", "B"}},
		"drops empty":     {Input: "A, , B", Expected: []string{"A", "B"}},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := SplitParticipants(test.Input)
			if !slices.Equal(test.Expected, got) {
				t.Errorf("SplitParticipants mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}

func TestReferencedParticipants(t *testing.T) {
	tests := map[string]struct {
		Lines    []string
		Expected []string
	}{
		"simple arrow": {
			Lines:    []string{"A -> B"},
			Expected: []string{"A", "B"},
		},
		"arrow with message": {
			Lines:    []string{"A -> B: hi"},
			Expected: []string{"A", "B"},
		},
		"dashed and double head": {
			Lines:    []string{"A --> B", "A ->> C"},
			Expected: []string{"A", "B", "C"},
		},
		"reverse arrow keeps appearance order": {
			Lines:    []string{"B <- A"},
			Expected: []string{"B", "A"},
		},
		"self message": {
			Lines:    []string{"A -> A"},
			Expected: []string{"A"},
		},
		"arrow without spaces": {
			Lines:    []string{"A->B"},
			Expected: []string{"A", "B"},
		},
		"activate deactivate destroy": {
			Lines:    []string{"activate X", "deactivate X", "destroy Y"},
			Expected: []string{"X", "Y"},
		},
		"create": {
			Lines:    []string{"create C", "A -> C"},
			Expected: []string{"C", "A"},
		},
		"note over single": {
			Lines:    []string{"note over A: x"},
			Expected: []string{"A"},
		},
		"note over multiple": {
			Lines:    []string{"note over A, B"},
			Expected: []string{"A", "B"},
		},
		"note right of": {
			Lines:    []string{"note right of A"},
			Expected: []string{"A"},
		},
		"ref over single": {
			Lines:    []string{"ref over A: ./sub.puml"},
			Expected: []string{"A"},
		},
		"ref over multiple": {
			Lines:    []string{"ref over A, B"},
			Expected: []string{"A", "B"},
		},
		"end ref is not a reference": {
			Lines:    []string{"end ref"},
			Expected: nil,
		},
		"quoted endpoints": {
			Lines:    []string{`"Svc One" -> B`},
			Expected: []string{"Svc One", "B"},
		},
		"declaration is not a reference": {
			Lines:    []string{"participant A"},
			Expected: nil,
		},
		"header lines ignored": {
			Lines:    []string{"@startuml", "@enduml"},
			Expected: nil,
		},
		"dedup keeps first appearance order": {
			Lines:    []string{"A -> B", "B -> A"},
			Expected: []string{"A", "B"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := ReferencedParticipants(test.Lines)
			if !slices.Equal(test.Expected, got) {
				t.Errorf("ReferencedParticipants mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}

func TestParseMessage(t *testing.T) {
	tests := map[string]struct {
		Line             string
		ExpectedSender   string
		ExpectedReceiver string
		ExpectedOK       bool
	}{
		"spaced arrow":         {Line: "a -> b", ExpectedSender: "a", ExpectedReceiver: "b", ExpectedOK: true},
		"arrow without spaces": {Line: "a->b", ExpectedSender: "a", ExpectedReceiver: "b", ExpectedOK: true},
		"reverse arrow":        {Line: "b <- a", ExpectedSender: "a", ExpectedReceiver: "b", ExpectedOK: true},
		"dashed arrow":         {Line: "a --> b", ExpectedSender: "a", ExpectedReceiver: "b", ExpectedOK: true},
		"double head":          {Line: "a ->> b", ExpectedSender: "a", ExpectedReceiver: "b", ExpectedOK: true},
		"lost message x":       {Line: "a ->x b", ExpectedSender: "a", ExpectedReceiver: "b", ExpectedOK: true},
		"circle head o":        {Line: "a ->o b", ExpectedSender: "a", ExpectedReceiver: "b", ExpectedOK: true},
		"reverse dashed":       {Line: "b <-- a", ExpectedSender: "a", ExpectedReceiver: "b", ExpectedOK: true},
		"reverse double head":  {Line: "b <<- a", ExpectedSender: "a", ExpectedReceiver: "b", ExpectedOK: true},

		"bidirectional has no direction": {Line: "a <-> b", ExpectedOK: false},
		"no arrow head":                  {Line: "a -- b", ExpectedOK: false},
		"declaration is not a message":   {Line: "participant A", ExpectedOK: false},
		"note is not a message":          {Line: "note over A", ExpectedOK: false},
		"empty line":                     {Line: "", ExpectedOK: false},

		"quoted endpoints":     {Line: `"Svc One" -> B`, ExpectedSender: "Svc One", ExpectedReceiver: "B", ExpectedOK: true},
		"self message":         {Line: "a -> a", ExpectedSender: "a", ExpectedReceiver: "a", ExpectedOK: true},
		"message text ignored": {Line: "a -> b: hello", ExpectedSender: "a", ExpectedReceiver: "b", ExpectedOK: true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			gotSender, gotReceiver, gotOK := ParseMessage(test.Line)
			if gotOK != test.ExpectedOK {
				t.Errorf("ok mismatch (-want +got):\n%s", cmp.Diff(test.ExpectedOK, gotOK))
			}
			if gotSender != test.ExpectedSender {
				t.Errorf("sender mismatch (-want +got):\n%s", cmp.Diff(test.ExpectedSender, gotSender))
			}
			if gotReceiver != test.ExpectedReceiver {
				t.Errorf("receiver mismatch (-want +got):\n%s", cmp.Diff(test.ExpectedReceiver, gotReceiver))
			}
		})
	}
}

func TestParseMessageParts(t *testing.T) {
	tests := map[string]struct {
		Line       string
		Expected   MessageParts
		ExpectedOK bool
	}{
		"spaced arrow": {
			Line:       "a -> b",
			Expected:   MessageParts{Indent: "", LeftRaw: "a", Mid: " -> ", RightRaw: "b", Rest: "", Sender: "a", Receiver: "b"},
			ExpectedOK: true,
		},
		"arrow without spaces": {
			Line:       "a->b",
			Expected:   MessageParts{LeftRaw: "a", Mid: "->", RightRaw: "b", Sender: "a", Receiver: "b"},
			ExpectedOK: true,
		},
		"indented with label": {
			Line:       "  a -> b: hi",
			Expected:   MessageParts{Indent: "  ", LeftRaw: "a", Mid: " -> ", RightRaw: "b", Rest: ": hi", Sender: "a", Receiver: "b"},
			ExpectedOK: true,
		},
		"reverse arrow": {
			Line:       "b <- a",
			Expected:   MessageParts{LeftRaw: "b", Mid: " <- ", RightRaw: "a", Sender: "a", Receiver: "b"},
			ExpectedOK: true,
		},
		"self message": {
			Line:       "a -> a",
			Expected:   MessageParts{LeftRaw: "a", Mid: " -> ", RightRaw: "a", Sender: "a", Receiver: "a"},
			ExpectedOK: true,
		},
		"activation suffix kept in rest": {
			Line:       "a -> b ++",
			Expected:   MessageParts{LeftRaw: "a", Mid: " -> ", RightRaw: "b", Rest: " ++", Sender: "a", Receiver: "b"},
			ExpectedOK: true,
		},
		"quoted endpoints": {
			Line:       `"Svc One" -> B`,
			Expected:   MessageParts{LeftRaw: `"Svc One"`, Mid: " -> ", RightRaw: "B", Sender: "Svc One", Receiver: "B"},
			ExpectedOK: true,
		},
		"bidirectional has no direction": {Line: "a <-> b", ExpectedOK: false},
		"no arrow head":                  {Line: "a -- b", ExpectedOK: false},
		"declaration is not a message":   {Line: "participant A", ExpectedOK: false},
		"empty line":                     {Line: "", ExpectedOK: false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, gotOK := ParseMessageParts(test.Line)
			if gotOK != test.ExpectedOK {
				t.Errorf("ok mismatch (-want +got):\n%s", cmp.Diff(test.ExpectedOK, gotOK))
			}
			if !gotOK {
				return
			}
			if got != test.Expected {
				t.Errorf("parts mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
			if rebuilt := got.Indent + got.LeftRaw + got.Mid + got.RightRaw + got.Rest; rebuilt != test.Line {
				t.Errorf("not lossless: got %q, want %q", rebuilt, test.Line)
			}
		})
	}
}

func TestElseLabel(t *testing.T) {
	tests := map[string]struct {
		Line          string
		ExpectedLabel string
		ExpectedOK    bool
	}{
		"bare else":             {Line: "else", ExpectedLabel: "", ExpectedOK: true},
		"else with condition":   {Line: "else not found", ExpectedLabel: "not found", ExpectedOK: true},
		"indented else":         {Line: "  else  retry  ", ExpectedLabel: "retry", ExpectedOK: true},
		"alt is not else":       {Line: "alt x", ExpectedOK: false},
		"elsewhere is not else": {Line: "elsewhere -> A", ExpectedOK: false},
		"message is not else":   {Line: "A -> B", ExpectedOK: false},
		"empty":                 {Line: "", ExpectedOK: false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			gotLabel, gotOK := ElseLabel(test.Line)
			if gotOK != test.ExpectedOK {
				t.Errorf("ok mismatch (-want +got):\n%s", cmp.Diff(test.ExpectedOK, gotOK))
			}
			if gotLabel != test.ExpectedLabel {
				t.Errorf("label mismatch (-want +got):\n%s", cmp.Diff(test.ExpectedLabel, gotLabel))
			}
		})
	}
}

func TestGroupStartLabel(t *testing.T) {
	tests := map[string]struct {
		Line            string
		ExpectedKeyword string
		ExpectedLabel   string
		ExpectedOK      bool
	}{
		"group with label": {
			Line:            "group Auth",
			ExpectedKeyword: "group",
			ExpectedLabel:   "Auth",
			ExpectedOK:      true,
		},
		"group with extra spaces": {
			Line:            "  group   Auth  ",
			ExpectedKeyword: "group",
			ExpectedLabel:   "Auth",
			ExpectedOK:      true,
		},
		"group without label": {
			Line:            "group",
			ExpectedKeyword: "group",
			ExpectedLabel:   "",
			ExpectedOK:      true,
		},
		"alt is an opener": {
			Line:            "alt success",
			ExpectedKeyword: "alt",
			ExpectedLabel:   "success",
			ExpectedOK:      true,
		},
		"loop with multi word label": {
			Line:            "loop every minute",
			ExpectedKeyword: "loop",
			ExpectedLabel:   "every minute",
			ExpectedOK:      true,
		},
		"message is not an opener": {
			Line:       "A -> B",
			ExpectedOK: false,
		},
		"participant is not an opener": {
			Line:       "participant A",
			ExpectedOK: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			gotKeyword, gotLabel, gotOK := GroupStartLabel(test.Line)
			if gotOK != test.ExpectedOK {
				t.Errorf("ok mismatch (-want +got):\n%s", cmp.Diff(test.ExpectedOK, gotOK))
			}
			if gotKeyword != test.ExpectedKeyword {
				t.Errorf("keyword mismatch (-want +got):\n%s", cmp.Diff(test.ExpectedKeyword, gotKeyword))
			}
			if gotLabel != test.ExpectedLabel {
				t.Errorf("label mismatch (-want +got):\n%s", cmp.Diff(test.ExpectedLabel, gotLabel))
			}
		})
	}
}

func TestIsBlockStart(t *testing.T) {
	tests := map[string]struct {
		Line     string
		Expected bool
	}{
		"group":       {Line: "group Auth", Expected: true},
		"alt":         {Line: "alt cond", Expected: true},
		"message":     {Line: "A -> B", Expected: false},
		"end":         {Line: "end", Expected: false},
		"participant": {Line: "participant A", Expected: false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := IsBlockStart(test.Line)
			if got != test.Expected {
				t.Errorf("IsBlockStart mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}

func TestIsBlockEnd(t *testing.T) {
	tests := map[string]struct {
		Line     string
		Expected bool
	}{
		"plain end":           {Line: "end", Expected: true},
		"end with spaces":     {Line: "  end  ", Expected: true},
		"end group":           {Line: "end group", Expected: true},
		"end box is neutral":  {Line: "end box", Expected: false},
		"end note is neutral": {Line: "end note", Expected: false},
		"message":             {Line: "A -> B", Expected: false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := IsBlockEnd(test.Line)
			if got != test.Expected {
				t.Errorf("IsBlockEnd mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}

func TestDeclaredParticipant(t *testing.T) {
	tests := map[string]struct {
		Line         string
		ExpectedName string
		ExpectedOK   bool
	}{
		"bare participant": {
			Line:         "participant A",
			ExpectedName: "A",
			ExpectedOK:   true,
		},
		"actor": {
			Line:         "actor User",
			ExpectedName: "User",
			ExpectedOK:   true,
		},
		"indented participant": {
			Line:         "  participant A",
			ExpectedName: "A",
			ExpectedOK:   true,
		},
		"quoted name with alias": {
			Line:         `participant "Long Name" as LN`,
			ExpectedName: "LN",
			ExpectedOK:   true,
		},
		"database quoted name with alias": {
			Line:         `database "DB" as db`,
			ExpectedName: "db",
			ExpectedOK:   true,
		},
		"quoted name without alias": {
			Line:         `participant "Svc"`,
			ExpectedName: "Svc",
			ExpectedOK:   true,
		},
		"not a declaration": {
			Line:         "A -> B: hi",
			ExpectedName: "",
			ExpectedOK:   false,
		},
		"message text contains keyword": {
			Line:         "A -> B: participant joins",
			ExpectedName: "",
			ExpectedOK:   false,
		},
		"keyword only without name": {
			Line:         "participant",
			ExpectedName: "",
			ExpectedOK:   false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			gotName, gotOK := DeclaredParticipant(test.Line)
			if gotName != test.ExpectedName {
				t.Errorf("name mismatch (-want +got):\n%s", cmp.Diff(test.ExpectedName, gotName))
			}
			if gotOK != test.ExpectedOK {
				t.Errorf("ok mismatch (-want +got):\n%s", cmp.Diff(test.ExpectedOK, gotOK))
			}
		})
	}
}

func TestIsNoteOpener(t *testing.T) {
	tests := map[string]struct {
		Input    string
		Expected bool
	}{
		"note left":                  {Input: "note left", Expected: true},
		"note over participants":     {Input: "note over Alice, Bob", Expected: true},
		"note right of":              {Input: "note right of Alice", Expected: true},
		"hnote":                      {Input: "hnote over Alice", Expected: true},
		"rnote":                      {Input: "rnote over Bob", Expected: true},
		"bare note":                  {Input: "note", Expected: true},
		"single line note with text": {Input: "note left: TODO something", Expected: true},
		"indented note":              {Input: "  note left", Expected: true},
		"uppercase NOTE":             {Input: "NOTE left", Expected: true},
		"note-prefixed identifier":   {Input: "notebook -> Alice: hi", Expected: false},
		"message arrow":              {Input: "Alice -> Bob: hello", Expected: false},
		"empty":                      {Input: "", Expected: false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := IsNoteOpener(test.Input)
			if got != test.Expected {
				t.Errorf("IsNoteOpener(%q) mismatch (-want +got):\n%s", test.Input, cmp.Diff(test.Expected, got))
			}
		})
	}
}

func TestIsNoteTerminator(t *testing.T) {
	tests := map[string]struct {
		Input    string
		Expected bool
	}{
		"end note":           {Input: "end note", Expected: true},
		"endnote":            {Input: "endnote", Expected: true},
		"endrnote":           {Input: "endrnote", Expected: true},
		"endhnote":           {Input: "endhnote", Expected: true},
		"end rnote":          {Input: "end rnote", Expected: true},
		"end hnote":          {Input: "end hnote", Expected: true},
		"indented end note":  {Input: "  end note  ", Expected: true},
		"uppercase END NOTE": {Input: "END NOTE", Expected: true},
		"note opener":        {Input: "note left", Expected: false},
		"empty":              {Input: "", Expected: false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := IsNoteTerminator(test.Input)
			if got != test.Expected {
				t.Errorf("IsNoteTerminator(%q) mismatch (-want +got):\n%s", test.Input, cmp.Diff(test.Expected, got))
			}
		})
	}
}

func TestIsRefOpener(t *testing.T) {
	tests := map[string]struct {
		Input    string
		Expected bool
	}{
		"ref over participants":     {Input: "ref over Alice, Bob", Expected: true},
		"ref over single":           {Input: "ref over Alice", Expected: true},
		"single line ref with path": {Input: "ref over Alice, Bob : ./auth.puml", Expected: true},
		"bare ref":                  {Input: "ref", Expected: true},
		"indented ref":              {Input: "  ref over Alice", Expected: true},
		"uppercase REF":             {Input: "REF over Alice", Expected: true},
		"ref-prefixed identifier":   {Input: "reference -> Alice: hi", Expected: false},
		"note opener":               {Input: "note over Alice", Expected: false},
		"message arrow":             {Input: "Alice -> Bob: hello", Expected: false},
		"empty":                     {Input: "", Expected: false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := IsRefOpener(test.Input)
			if got != test.Expected {
				t.Errorf("IsRefOpener(%q) mismatch (-want +got):\n%s", test.Input, cmp.Diff(test.Expected, got))
			}
		})
	}
}

func TestIsRefTerminator(t *testing.T) {
	tests := map[string]struct {
		Input    string
		Expected bool
	}{
		"end ref":           {Input: "end ref", Expected: true},
		"endref":            {Input: "endref", Expected: true},
		"indented end ref":  {Input: "  end ref  ", Expected: true},
		"uppercase END REF": {Input: "END REF", Expected: true},
		"ref opener":        {Input: "ref over Alice", Expected: false},
		"end note":          {Input: "end note", Expected: false},
		"bare end":          {Input: "end", Expected: false},
		"empty":             {Input: "", Expected: false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := IsRefTerminator(test.Input)
			if got != test.Expected {
				t.Errorf("IsRefTerminator(%q) mismatch (-want +got):\n%s", test.Input, cmp.Diff(test.Expected, got))
			}
		})
	}
}

func TestIsBoxStart(t *testing.T) {
	tests := map[string]struct {
		Input    string
		Expected bool
	}{
		"box with label":          {Input: `box "Front"`, Expected: true},
		"bare box":                {Input: "box", Expected: true},
		"box with color":          {Input: `box "Front" #LightBlue`, Expected: true},
		"indented box":            {Input: `  box "Front"`, Expected: true},
		"uppercase BOX":           {Input: "BOX", Expected: true},
		"end box is not a start":  {Input: "end box", Expected: false},
		"box-prefixed identifier": {Input: "boxer -> Alice: hi", Expected: false},
		"empty":                   {Input: "", Expected: false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := IsBoxStart(test.Input)
			if got != test.Expected {
				t.Errorf("IsBoxStart(%q) mismatch (-want +got):\n%s", test.Input, cmp.Diff(test.Expected, got))
			}
		})
	}
}

func TestBoxLabel(t *testing.T) {
	tests := map[string]struct {
		Input    string
		Expected string
		OK       bool
	}{
		"quoted label":            {Input: `box "Front"`, Expected: "Front", OK: true},
		"quoted label with color": {Input: `box "Ops" #888888`, Expected: "Ops", OK: true},
		"unquoted label":          {Input: "box Front", Expected: "Front", OK: true},
		"indented box":            {Input: `  box "Front"  `, Expected: "Front", OK: true},
		"uppercase BOX":           {Input: `BOX "Front"`, Expected: "Front", OK: true},
		"bare box":                {Input: "box", Expected: "", OK: false},
		"color only":              {Input: "box #AliceBlue", Expected: "", OK: false},
		"quoted color is a label": {Input: `box "#888888"`, Expected: "#888888", OK: true},
		"end box is not a start":  {Input: "end box", Expected: "", OK: false},
		"box-prefixed identifier": {Input: "boxer -> Alice: hi", Expected: "", OK: false},
		"empty":                   {Input: "", Expected: "", OK: false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, ok := BoxLabel(test.Input)
			if got != test.Expected || ok != test.OK {
				t.Errorf("BoxLabel(%q) mismatch (-want +got):\n%s", test.Input,
					cmp.Diff([]any{test.Expected, test.OK}, []any{got, ok}))
			}
		})
	}
}

func TestIsBoxEnd(t *testing.T) {
	tests := map[string]struct {
		Input    string
		Expected bool
	}{
		"end box":           {Input: "end box", Expected: true},
		"endbox":            {Input: "endbox", Expected: true},
		"indented end box":  {Input: "  end box  ", Expected: true},
		"uppercase END BOX": {Input: "END BOX", Expected: true},
		"box opener":        {Input: `box "Front"`, Expected: false},
		"bare end":          {Input: "end", Expected: false},
		"empty":             {Input: "", Expected: false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := IsBoxEnd(test.Input)
			if got != test.Expected {
				t.Errorf("IsBoxEnd(%q) mismatch (-want +got):\n%s", test.Input, cmp.Diff(test.Expected, got))
			}
		})
	}
}
