package lines

import "strings"

func Split(src string) (ls []string, trailingNewline bool) {
	if strings.HasSuffix(src, "\n") {
		return strings.Split(strings.TrimSuffix(src, "\n"), "\n"), true
	}
	return strings.Split(src, "\n"), false
}

func Join(ls []string, trailingNewline bool) string {
	out := strings.Join(ls, "\n")
	if trailingNewline && out != "" {
		out += "\n"
	}
	return out
}

func FirstWithPrefix(ls []string, prefix string) (string, bool) {
	for _, line := range ls {
		if strings.HasPrefix(strings.TrimSpace(line), prefix) {
			return line, true
		}
	}
	return "", false
}

func LastWithPrefix(ls []string, prefix string) (string, bool) {
	for i := len(ls) - 1; i >= 0; i-- {
		if strings.HasPrefix(strings.TrimSpace(ls[i]), prefix) {
			return ls[i], true
		}
	}
	return "", false
}

func MatchingEnd(ls []string, start int, isStart, isEnd func(string) bool) int {
	depth := 0
	for i := start; i < len(ls); i++ {
		switch {
		case isStart(ls[i]):
			depth++
		case isEnd(ls[i]):
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func LeadingIndent(s string) string {
	return s[:len(s)-len(strings.TrimLeft(s, " \t"))]
}

func Indent(text, indent string) string {
	if indent == "" || text == "" {
		return text
	}
	ls := strings.Split(text, "\n")
	for i, l := range ls {
		if strings.TrimSpace(l) != "" {
			ls[i] = indent + l
		}
	}
	return strings.Join(ls, "\n")
}
