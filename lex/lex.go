package lex

import (
	"strings"
	"unicode"
)

func TokenizeQuoted(s string) []string {
	var tokens []string
	var b strings.Builder
	inQuote := false
	flush := func() {
		if b.Len() > 0 {
			tokens = append(tokens, b.String())
			b.Reset()
		}
	}
	for _, r := range s {
		switch {
		case r == '"':
			b.WriteRune(r)
			if inQuote {
				inQuote = false
				flush()
			} else {
				inQuote = true
			}
		case !inQuote && unicode.IsSpace(r):
			flush()
		default:
			b.WriteRune(r)
		}
	}
	flush()
	return tokens
}

func Unquote(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

func IndexOutsideQuotes(s string, sep byte) int {
	inQuote := false
	for i := 0; i < len(s); i++ {
		switch {
		case !inQuote && s[i] == sep:
			return i
		case s[i] == '"':
			inQuote = !inQuote
		}
	}
	return -1
}
