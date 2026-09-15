package todo

import (
	"regexp"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestExtract(t *testing.T) {
	tests := map[string]struct {
		Input    string
		Regexp   string
		Expected []Todo
	}{
		"empty input": {
			Input:    "",
			Expected: nil,
		},
		"single comment todo": {
			Input:    "'TODO: handle error case\n",
			Expected: []Todo{{Line: 1, Text: "TODO: handle error case"}},
		},
		"indented comment todo": {
			Input:    "  ' TODO refactor this\n",
			Expected: []Todo{{Line: 1, Text: "TODO refactor this"}},
		},
		"non-todo comment ignored": {
			Input:    "' just a normal comment\n",
			Expected: nil,
		},
		"comment todo among code": {
			Input:    "@startuml\nAlice -> Bob: hello\n'TODO: add reply\n@enduml\n",
			Expected: []Todo{{Line: 3, Text: "TODO: add reply"}},
		},
		"fixme matched with custom regexp": {
			Input:    "'FIXME: fix this\n",
			Regexp:   "(TODO|FIXME)",
			Expected: []Todo{{Line: 1, Text: "FIXME: fix this"}},
		},
		"fixme not matched with default regexp": {
			Input:    "'FIXME: fix this\n",
			Expected: nil,
		},
		"case insensitive with inline flag": {
			Input:    "' ToDo mixed case\n",
			Regexp:   "(?i)todo",
			Expected: []Todo{{Line: 1, Text: "ToDo mixed case"}},
		},
		"mixed case not matched with default regexp": {
			Input:    "' ToDo mixed case\n",
			Expected: nil,
		},
		"note left single line": {
			Input:    "note left: TODO something\n",
			Expected: []Todo{{Line: 1, Text: "TODO something"}},
		},
		"note right of single line": {
			Input:    "note right of Alice: TODO x\n",
			Expected: []Todo{{Line: 1, Text: "TODO x"}},
		},
		"note over participants single line": {
			Input:    "note over Alice, Bob : TODO y\n",
			Expected: []Todo{{Line: 1, Text: "TODO y"}},
		},
		"hnote single line": {
			Input:    "hnote over Alice: TODO h\n",
			Expected: []Todo{{Line: 1, Text: "TODO h"}},
		},
		"rnote single line": {
			Input:    "rnote over Bob: TODO r\n",
			Expected: []Todo{{Line: 1, Text: "TODO r"}},
		},
		"ref single line": {
			Input:    "ref over Alice, Bob : TODO write auth.puml\n",
			Expected: []Todo{{Line: 1, Text: "TODO write auth.puml"}},
		},
		"ref multiline block": {
			Input:    "ref over Alice, Bob\n  TODO write auth.puml\nend ref\n",
			Expected: []Todo{{Line: 1, Text: "TODO write auth.puml"}},
		},
		"ref multiline block without todo": {
			Input:    "ref over Alice, Bob\n  ./auth.puml\nend ref\n",
			Expected: nil,
		},
		"ref-prefixed identifier is not a ref": {
			Input:    "reference -> Alice: TODO z\n",
			Expected: nil,
		},
		"ref body arrow is not mistaken for code": {
			Input:    "ref over Alice, Bob\n  Alice -> Bob: TODO m\nend ref\n",
			Expected: []Todo{{Line: 1, Text: "Alice -> Bob: TODO m"}},
		},
		"note single line without todo": {
			Input:    "note left: just an explanation\n",
			Expected: nil,
		},
		"note-prefixed identifier is not a note": {
			Input:    "notebook -> Alice: TODO z\n",
			Expected: nil,
		},
		"note block single body line": {
			Input:    "note left\nTODO write this\nend note\n",
			Expected: []Todo{{Line: 1, Text: "TODO write this"}},
		},
		"note block whole body when one line matches": {
			Input:    "note over Alice\nfirst line\nTODO do the thing\nlast line\nend note\n",
			Expected: []Todo{{Line: 1, Text: "first line\nTODO do the thing\nlast line"}},
		},
		"note block without todo": {
			Input:    "note left\njust an explanation\nend note\n",
			Expected: nil,
		},
		"note block endnote spelling": {
			Input:    "note left\nTODO x\nendnote\n",
			Expected: []Todo{{Line: 1, Text: "TODO x"}},
		},
		"rnote block": {
			Input:    "rnote over Bob\nTODO r\nendrnote\n",
			Expected: []Todo{{Line: 1, Text: "TODO r"}},
		},
		"hnote block": {
			Input:    "hnote over Bob\nTODO h\nendhnote\n",
			Expected: []Todo{{Line: 1, Text: "TODO h"}},
		},
		"block comment single line": {
			Input:    "/' TODO in block '/\n",
			Expected: []Todo{{Line: 1, Text: "TODO in block"}},
		},
		"block comment single line without todo": {
			Input:    "/' just a block note '/\n",
			Expected: nil,
		},
		"block comment multiline whole body": {
			Input:    "/'\nsome text\nTODO inside block\nmore text\n'/\n",
			Expected: []Todo{{Line: 1, Text: "some text\nTODO inside block\nmore text"}},
		},
		"block comment multiline inline opener text": {
			Input:    "/' TODO on opener\nmore\n'/\n",
			Expected: []Todo{{Line: 1, Text: "TODO on opener\nmore"}},
		},
		"block comment multiline without todo": {
			Input:    "/'\njust text\nmore text\n'/\n",
			Expected: nil,
		},
		"mixed sources in source order": {
			Input: "@startuml\n" +
				"'TODO: first comment\n" +
				"Alice -> Bob: hi\n" +
				"note left: TODO second\n" +
				"note over Bob\n" +
				"TODO third in block\n" +
				"end note\n" +
				"/'\n" +
				"TODO fourth\n" +
				"'/\n" +
				"@enduml\n",
			Expected: []Todo{
				{Line: 2, Text: "TODO: first comment"},
				{Line: 4, Text: "TODO second"},
				{Line: 5, Text: "TODO third in block"},
				{Line: 8, Text: "TODO fourth"},
			},
		},
		"no trailing newline": {
			Input:    "'TODO: no newline at eof",
			Expected: []Todo{{Line: 1, Text: "TODO: no newline at eof"}},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			pattern := test.Regexp
			if pattern == "" {
				pattern = "(TODO|todo)"
			}
			re := regexp.MustCompile(pattern)

			got := Extract([]byte(test.Input), re)

			if !slices.Equal(test.Expected, got) {
				t.Errorf("Extract mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}
