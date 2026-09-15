package cli_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/Kuniwak/puml-tools/cli"
)

func TestOpenInput(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "input.txt")
	if err := os.WriteFile(path, []byte("from file"), 0o600); err != nil {
		t.Fatalf("WriteFile: %s", err)
	}
	missing := filepath.Join(dir, "missing.txt")

	tests := []struct {
		Name        string
		InputName   string
		Stdin       string
		Expected    string
		ExpectedErr bool
	}{
		{
			Name:      "- returns standard input",
			InputName: "-",
			Stdin:     "from stdin",
			Expected:  "from stdin",
		},
		{
			Name:      "a path opens that file",
			InputName: path,
			Stdin:     "from stdin",
			Expected:  "from file",
		},
		{
			Name:        "a nonexistent path is an error",
			InputName:   missing,
			Stdin:       "from stdin",
			ExpectedErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			spy := cli.SpyProcInout(test.Stdin)
			r, err := cli.OpenInput(test.InputName, spy.NewProcInout())
			if test.ExpectedErr {
				if err == nil {
					t.Fatalf("OpenInput(%q) should fail, but got no error", test.InputName)
				}
				return
			}
			if err != nil {
				t.Fatalf("OpenInput(%q) should succeed, but got error: %s", test.InputName, err)
			}
			defer r.Close()

			b, err := io.ReadAll(r)
			if err != nil {
				t.Fatalf("ReadAll: %s", err)
			}
			if got := string(b); got != test.Expected {
				t.Errorf("content mismatch: want %q, got %q", test.Expected, got)
			}
		})
	}
}
