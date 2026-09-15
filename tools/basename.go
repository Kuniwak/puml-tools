package tools

import (
	"path/filepath"
	"strings"
)

func Basename(path string) string {
	name := filepath.Base(path)
	name = strings.TrimSuffix(name, ".png")
	return strings.TrimSuffix(name, ".puml")
}
