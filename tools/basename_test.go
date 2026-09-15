package tools

import (
	"testing"
)

func TestBasename(t *testing.T) {
	tests := map[string]struct {
		Path     string
		Expected string
	}{
		"puml file":                  {Path: "seq.puml", Expected: "seq"},
		"path is dropped":            {Path: "./docs/design/seq.puml", Expected: "seq"},
		"png of puml drops both":     {Path: "docs/seq.puml.png", Expected: "seq"},
		"png alone drops png":        {Path: "docs/seq.png", Expected: "seq"},
		"unknown extension is kept":  {Path: "docs/seq.txt", Expected: "seq.txt"},
		"no extension is kept as is": {Path: "docs/seq", Expected: "seq"},
		"dots in the name are kept":  {Path: "docs/seq.v2.puml", Expected: "seq.v2"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := Basename(test.Path); got != test.Expected {
				t.Errorf("Basename(%q) = %q, want %q", test.Path, got, test.Expected)
			}
		})
	}
}
