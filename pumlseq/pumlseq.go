package pumlseq

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/Kuniwak/puml-tools/lex"
	"github.com/Kuniwak/puml-tools/sets"
)

var arrowRe = regexp.MustCompile(`^\s*("[^"]*"|[^\s"]+?)\s*([-\\/<>ox]*[-<>\\/][-\\/<>ox]*)\s*("[^"]*"|[^\s":]+)`)

func ParseMessage(line string) (sender, receiver string, ok bool) {
	m := arrowRe.FindStringSubmatch(strings.TrimSpace(line))
	if m == nil {
		return "", "", false
	}
	left, right := lex.Unquote(m[1]), lex.Unquote(m[3])
	hasGt, hasLt := strings.Contains(m[2], ">"), strings.Contains(m[2], "<")
	switch {
	case hasGt && !hasLt:
		return left, right, true
	case hasLt && !hasGt:
		return right, left, true
	default:
		return "", "", false
	}
}

type MessageParts struct {
	Indent   string
	LeftRaw  string
	Mid      string
	RightRaw string
	Rest     string
	Sender   string
	Receiver string
}

func ParseMessageParts(line string) (MessageParts, bool) {
	idx := arrowRe.FindStringSubmatchIndex(line)
	if idx == nil {
		return MessageParts{}, false
	}
	connector := line[idx[4]:idx[5]]
	hasGt, hasLt := strings.Contains(connector, ">"), strings.Contains(connector, "<")
	left, right := lex.Unquote(line[idx[2]:idx[3]]), lex.Unquote(line[idx[6]:idx[7]])
	var sender, receiver string
	switch {
	case hasGt && !hasLt:
		sender, receiver = left, right
	case hasLt && !hasGt:
		sender, receiver = right, left
	default:
		return MessageParts{}, false
	}
	return MessageParts{
		Indent:   line[:idx[2]],
		LeftRaw:  line[idx[2]:idx[3]],
		Mid:      line[idx[3]:idx[6]],
		RightRaw: line[idx[6]:idx[7]],
		Rest:     line[idx[7]:],
		Sender:   sender,
		Receiver: receiver,
	}, true
}

func ElseLabel(line string) (label string, ok bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return "", false
	}
	first := trimmed
	rest := ""
	if i := strings.IndexFunc(trimmed, unicode.IsSpace); i >= 0 {
		first = trimmed[:i]
		rest = strings.TrimSpace(trimmed[i+1:])
	}
	if first != "else" {
		return "", false
	}
	return rest, true
}

func ReferencedParticipants(ls []string) []string {
	var result []string
	seen := sets.New()
	add := func(name string) {
		name = strings.TrimSpace(name)
		if name == "" || seen.Contains(name) {
			return
		}
		seen.Add(name)
		result = append(result, name)
	}

	for _, line := range ls {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if _, ok := DeclaredParticipant(trimmed); ok {
			continue
		}
		if m := arrowRe.FindStringSubmatch(trimmed); m != nil {
			add(lex.Unquote(m[1]))
			add(lex.Unquote(m[3]))
			continue
		}
		fields := strings.Fields(trimmed)
		switch fields[0] {
		case "activate", "deactivate", "destroy", "create":
			rest := strings.TrimSpace(trimmed[len(fields[0]):])
			if name, ok := DeclaredParticipant(rest); ok {
				add(name)
			} else if len(fields) >= 2 {
				add(lex.Unquote(fields[1]))
			}
		case "note", "rnote", "hnote", "ref":
			for _, n := range NoteTargets(trimmed) {
				add(n)
			}
		}
	}
	return result
}

func NoteTargets(line string) []string {
	body := line
	if i := lex.IndexOutsideQuotes(body, ':'); i >= 0 {
		body = body[:i]
	}
	if idx := strings.Index(body, " over "); idx >= 0 {
		return SplitParticipants(body[idx+len(" over "):])
	}
	if idx := strings.Index(body, " of "); idx >= 0 {
		return SplitParticipants(body[idx+len(" of "):])
	}
	return nil
}

func SplitParticipants(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if p := lex.Unquote(strings.TrimSpace(part)); p != "" {
			out = append(out, p)
		}
	}
	return out
}

var participantKeywords = sets.New(
	"participant",
	"actor",
	"boundary",
	"control",
	"entity",
	"database",
	"collections",
	"queue",
)

type MatchFunc func(label string) bool

func Exact(label string) MatchFunc {
	return func(l string) bool { return l == label }
}

var blockStartKeywords = sets.New(
	"group",
	"alt",
	"opt",
	"loop",
	"par",
	"break",
	"critical",
)

func GroupStartLabel(line string) (keyword string, label string, ok bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return "", "", false
	}
	first := trimmed
	rest := ""
	if i := strings.IndexFunc(trimmed, unicode.IsSpace); i >= 0 {
		first = trimmed[:i]
		rest = strings.TrimSpace(trimmed[i+1:])
	}
	if !blockStartKeywords.Contains(first) {
		return "", "", false
	}
	return first, rest, true
}

func IsBlockStart(line string) bool {
	_, _, ok := GroupStartLabel(line)
	return ok
}

func IsBlockEnd(line string) bool {
	fields := strings.Fields(line)
	if len(fields) == 0 || fields[0] != "end" {
		return false
	}
	if len(fields) == 1 {
		return true
	}
	return blockStartKeywords.Contains(fields[1])
}

func DeclaredParticipant(line string) (name string, ok bool) {
	tokens := lex.TokenizeQuoted(strings.TrimSpace(line))
	if len(tokens) < 2 || !participantKeywords.Contains(tokens[0]) {
		return "", false
	}
	for i := 1; i+1 < len(tokens); i++ {
		if tokens[i] == "as" {
			return lex.Unquote(tokens[i+1]), true
		}
	}
	return lex.Unquote(tokens[1]), true
}

func IsNoteOpener(line string) bool {
	for _, keyword := range []string{"note", "hnote", "rnote"} {
		if hasKeyword(line, keyword) {
			return true
		}
	}
	return false
}

func IsNoteTerminator(line string) bool {
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "end note", "endnote", "end rnote", "endrnote", "end hnote", "endhnote":
		return true
	}
	return false
}

func IsRefOpener(line string) bool {
	return hasKeyword(line, "ref")
}

func hasKeyword(line, keyword string) bool {
	lower := strings.ToLower(strings.TrimSpace(line))
	if !strings.HasPrefix(lower, keyword) {
		return false
	}
	rest := lower[len(keyword):]
	return rest == "" || rest[0] == ' ' || rest[0] == '\t'
}

func IsBoxStart(line string) bool {
	return hasKeyword(line, "box")
}

func BoxLabel(line string) (string, bool) {
	if !IsBoxStart(line) {
		return "", false
	}
	tokens := lex.TokenizeQuoted(strings.TrimSpace(line))
	if len(tokens) < 2 || strings.HasPrefix(tokens[1], "#") {
		return "", false
	}
	return lex.Unquote(tokens[1]), true
}

func IsBoxEnd(line string) bool {
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "end box", "endbox":
		return true
	}
	return false
}

func IsRefTerminator(line string) bool {
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "end ref", "endref":
		return true
	}
	return false
}
