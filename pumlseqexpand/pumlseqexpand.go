package pumlseqexpand

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var schemeRe = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9+.-]*://`)

var volumeRe = regexp.MustCompile(`^[A-Za-z]:`)

func ParsePath(label string) (string, error) {
	path := strings.TrimSpace(label)
	switch {
	case path == "":
		return "", fmt.Errorf("ref label is empty; specify the reference target as a relative path")
	case strings.ContainsAny(path, "\n\r"):
		return "", fmt.Errorf("ref label spans multiple lines: %q; specify only the reference target's relative path", label)
	case schemeRe.MatchString(path):
		return "", fmt.Errorf("ref label is a URL: %q; the reference target must be a relative path", path)
	case strings.HasPrefix(path, "~"):
		return "", fmt.Errorf("ref label starts from the home directory: %q; the reference target must be a relative path", path)
	case volumeRe.MatchString(path):
		return "", fmt.Errorf("ref label includes a volume name: %q; the reference target must be a relative path", path)
	case filepath.IsAbs(path) || strings.HasPrefix(path, "/") || strings.HasPrefix(path, `\`):
		return "", fmt.Errorf("ref label is an absolute path: %q; the reference target must be a relative path", path)
	case strings.ContainsAny(path, " \t"):
		return "", fmt.Errorf("ref label is not a relative path: %q; specify the reference target as a relative path", path)
	}
	return path, nil
}
