package lex

import (
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestTokenizeQuoted(t *testing.T) {
	tests := map[string]struct {
		Input    string
		Expected []string
	}{
		"simple": {
			Input:    "participant A",
			Expected: []string{"participant", "A"},
		},
		"quoted with spaces and alias": {
			Input:    `participant "Long Name" as LN`,
			Expected: []string{"participant", `"Long Name"`, "as", "LN"},
		},
		"quoted only": {
			Input:    `"Svc"`,
			Expected: []string{`"Svc"`},
		},
		"extra spaces": {
			Input:    "  a   b  ",
			Expected: []string{"a", "b"},
		},
		"empty": {
			Input:    "",
			Expected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := TokenizeQuoted(test.Input)
			if !slices.Equal(test.Expected, got) {
				t.Errorf("TokenizeQuoted mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}

func TestUnquote(t *testing.T) {
	tests := map[string]struct {
		Input    string
		Expected string
	}{
		"quoted":            {Input: `"abc"`, Expected: "abc"},
		"unquoted":          {Input: "abc", Expected: "abc"},
		"quoted with space": {Input: `"a b"`, Expected: "a b"},
		"empty quotes":      {Input: `""`, Expected: ""},
		"single quote char": {Input: `"`, Expected: `"`},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := Unquote(test.Input); got != test.Expected {
				t.Errorf("Unquote(%q) = %q, want %q", test.Input, got, test.Expected)
			}
		})
	}
}

func TestIndexOutsideQuotes(t *testing.T) {
	tests := map[string]struct {
		Input    string
		Sep      byte
		Expected int
	}{
		"absent":                     {Input: "abc", Sep: ':', Expected: -1},
		"outside quotes":             {Input: "a:b", Sep: ':', Expected: 1},
		"inside quotes is skipped":   {Input: `"a:b"`, Sep: ':', Expected: -1},
		"after closing quote":        {Input: `"a:b":c`, Sep: ':', Expected: 5},
		"first outside occurrence":   {Input: "a:b:c", Sep: ':', Expected: 1},
		"unterminated quote skips":   {Input: `"a:b`, Sep: ':', Expected: -1},
		"quote char itself is found": {Input: `a"b`, Sep: '"', Expected: 1},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := IndexOutsideQuotes(test.Input, test.Sep); got != test.Expected {
				t.Errorf("IndexOutsideQuotes(%q, %q) = %d, want %d", test.Input, test.Sep, got, test.Expected)
			}
		})
	}
}
