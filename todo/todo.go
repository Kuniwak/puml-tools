package todo

import (
	"regexp"
	"strings"

	"github.com/Kuniwak/puml-tools/pumlseq"
)

type Todo struct {
	Line int
	Text string
}

func Extract(src []byte, re *regexp.Regexp) []Todo {
	var todos []Todo

	var (
		opaqueTerm    func(string) bool
		opaqueStart   int
		opaqueBody    []string
		opaqueMatched bool

		inBlockComment bool
		blockStart     int
		blockBody      []string
		blockMatched   bool
	)

	for i, line := range strings.Split(string(src), "\n") {
		trimmed := strings.TrimSpace(strings.TrimRight(line, "\r"))

		if opaqueTerm != nil {
			if opaqueTerm(trimmed) {
				if opaqueMatched {
					todos = append(todos, Todo{Line: opaqueStart, Text: strings.TrimSpace(strings.Join(opaqueBody, "\n"))})
				}
				opaqueTerm, opaqueBody, opaqueMatched = nil, nil, false
				continue
			}
			opaqueBody = append(opaqueBody, trimmed)
			if re.MatchString(trimmed) {
				opaqueMatched = true
			}
			continue
		}

		if inBlockComment {
			if idx := strings.Index(trimmed, "'/"); idx >= 0 {
				if frag := strings.TrimSpace(trimmed[:idx]); frag != "" {
					blockBody = append(blockBody, frag)
					if re.MatchString(frag) {
						blockMatched = true
					}
				}
				if blockMatched {
					todos = append(todos, Todo{Line: blockStart, Text: strings.TrimSpace(strings.Join(blockBody, "\n"))})
				}
				inBlockComment, blockBody, blockMatched = false, nil, false
				continue
			}
			if trimmed != "" {
				blockBody = append(blockBody, trimmed)
				if re.MatchString(trimmed) {
					blockMatched = true
				}
			}
			continue
		}

		switch {
		case strings.HasPrefix(trimmed, "/'"):
			rest := trimmed[len("/'"):]
			if idx := strings.Index(rest, "'/"); idx >= 0 {
				content := strings.TrimSpace(rest[:idx])
				if re.MatchString(content) {
					todos = append(todos, Todo{Line: i + 1, Text: content})
				}
				continue
			}
			inBlockComment, blockStart, blockBody, blockMatched = true, i+1, nil, false
			if frag := strings.TrimSpace(rest); frag != "" {
				blockBody = append(blockBody, frag)
				if re.MatchString(frag) {
					blockMatched = true
				}
			}

		case strings.HasPrefix(trimmed, "'"):
			body := trimmed[len("'"):]
			if re.MatchString(body) {
				todos = append(todos, Todo{Line: i + 1, Text: strings.TrimSpace(body)})
			}

		case pumlseq.IsNoteOpener(trimmed):
			if idx := strings.Index(trimmed, ":"); idx >= 0 {
				body := trimmed[idx+1:]
				if re.MatchString(body) {
					todos = append(todos, Todo{Line: i + 1, Text: strings.TrimSpace(body)})
				}
			} else {
				opaqueTerm, opaqueStart, opaqueBody, opaqueMatched = pumlseq.IsNoteTerminator, i+1, nil, false
			}

		case pumlseq.IsRefOpener(trimmed):
			if idx := strings.Index(trimmed, ":"); idx >= 0 {
				body := trimmed[idx+1:]
				if re.MatchString(body) {
					todos = append(todos, Todo{Line: i + 1, Text: strings.TrimSpace(body)})
				}
			} else {
				opaqueTerm, opaqueStart, opaqueBody, opaqueMatched = pumlseq.IsRefTerminator, i+1, nil, false
			}
		}
	}
	return todos
}
