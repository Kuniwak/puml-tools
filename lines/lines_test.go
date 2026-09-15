package lines

import (
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSplit(t *testing.T) {
	tests := map[string]struct {
		Src             string
		ExpectedLines   []string
		ExpectedTrailNL bool
	}{
		"with trailing newline": {
			Src:             "a\nb\n",
			ExpectedLines:   []string{"a", "b"},
			ExpectedTrailNL: true,
		},
		"without trailing newline": {
			Src:             "a\nb",
			ExpectedLines:   []string{"a", "b"},
			ExpectedTrailNL: false,
		},
		"single line with newline": {
			Src:             "a\n",
			ExpectedLines:   []string{"a"},
			ExpectedTrailNL: true,
		},
		"empty string": {
			Src:             "",
			ExpectedLines:   []string{""},
			ExpectedTrailNL: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			gotLines, gotTrailNL := Split(test.Src)
			if !slices.Equal(test.ExpectedLines, gotLines) {
				t.Errorf("lines mismatch (-want +got):\n%s", cmp.Diff(test.ExpectedLines, gotLines))
			}
			if gotTrailNL != test.ExpectedTrailNL {
				t.Errorf("trailingNewline = %v, want %v", gotTrailNL, test.ExpectedTrailNL)
			}
		})
	}
}

func TestJoin(t *testing.T) {
	tests := map[string]struct {
		Lines    []string
		TrailNL  bool
		Expected string
	}{
		"with trailing newline":    {Lines: []string{"a", "b"}, TrailNL: true, Expected: "a\nb\n"},
		"without trailing newline": {Lines: []string{"a", "b"}, TrailNL: false, Expected: "a\nb"},
		"empty stays empty":        {Lines: []string{""}, TrailNL: true, Expected: ""},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := Join(test.Lines, test.TrailNL)
			if got != test.Expected {
				t.Errorf("Join mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}

func TestFirstWithPrefix(t *testing.T) {
	lines := []string{"x", "  @startuml Seq", "@startuml again", "y"}
	got, ok := FirstWithPrefix(lines, "@startuml")
	if !ok || got != "  @startuml Seq" {
		t.Errorf("FirstWithPrefix = (%q, %v), want (%q, true)", got, ok, "  @startuml Seq")
	}
	if _, ok := FirstWithPrefix(lines, "@enduml"); ok {
		t.Errorf("FirstWithPrefix found non-existent prefix")
	}
}

func TestLastWithPrefix(t *testing.T) {
	lines := []string{"@enduml first", "y", "@enduml last"}
	got, ok := LastWithPrefix(lines, "@enduml")
	if !ok || got != "@enduml last" {
		t.Errorf("LastWithPrefix = (%q, %v), want (%q, true)", got, ok, "@enduml last")
	}
}

func TestMatchingEnd(t *testing.T) {
	isStart := func(s string) bool { return s == "(" }
	isEnd := func(s string) bool { return s == ")" }

	tests := map[string]struct {
		Lines    []string
		Start    int
		Expected int
	}{
		"flat":             {Lines: []string{"(", "x", ")"}, Start: 0, Expected: 2},
		"nested outer":     {Lines: []string{"(", "(", ")", ")"}, Start: 0, Expected: 3},
		"nested inner":     {Lines: []string{"(", "(", ")", ")"}, Start: 1, Expected: 2},
		"no closing is -1": {Lines: []string{"(", "x"}, Start: 0, Expected: -1},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := MatchingEnd(test.Lines, test.Start, isStart, isEnd)
			if got != test.Expected {
				t.Errorf("MatchingEnd = %d, want %d", got, test.Expected)
			}
		})
	}
}

func TestLeadingIndent(t *testing.T) {
	tests := map[string]struct {
		Input    string
		Expected string
	}{
		"no indent":                      {Input: "a -> b", Expected: ""},
		"spaces":                         {Input: "  a -> b", Expected: "  "},
		"tab":                            {Input: "\ta -> b", Expected: "\t"},
		"mixed spaces and tabs":          {Input: " \t a", Expected: " \t "},
		"whitespace-only line":           {Input: "   ", Expected: "   "},
		"empty":                          {Input: "", Expected: ""},
		"inline whitespace is not taken": {Input: "a  b", Expected: ""},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := LeadingIndent(test.Input); got != test.Expected {
				t.Errorf("LeadingIndent(%q) = %q, want %q", test.Input, got, test.Expected)
			}
		})
	}
}

func TestIndent(t *testing.T) {
	tests := map[string]struct {
		Text     string
		Indent   string
		Expected string
	}{
		"adds to every line":          {Text: "a\nb\n", Indent: "  ", Expected: "  a\n  b\n"},
		"does not add to blank lines": {Text: "a\n\nb\n", Indent: "  ", Expected: "  a\n\n  b\n"},
		"preserves relative nesting":  {Text: "a\n  b\n", Indent: "  ", Expected: "  a\n    b\n"},
		"no-op when indent is empty":  {Text: "a\nb\n", Indent: "", Expected: "a\nb\n"},
		"no-op when text is empty":    {Text: "", Indent: "  ", Expected: ""},
		"no trailing newline":         {Text: "a\nb", Indent: "  ", Expected: "  a\n  b"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := Indent(test.Text, test.Indent); got != test.Expected {
				t.Errorf("Indent(%q, %q) = %q, want %q", test.Text, test.Indent, got, test.Expected)
			}
		})
	}
}
