package cli_test

import (
	"maps"
	"os"
	"path/filepath"
	"testing"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/google/go-cmp/cmp"
)

func TestReadDirContents(t *testing.T) {
	tests := map[string]struct {
		Files    map[string]string
		Expected map[string]string
	}{
		"returns nil when the directory does not exist": {
			Files:    nil,
			Expected: nil,
		},
		"returns nil for an empty directory": {
			Files:    map[string]string{},
			Expected: nil,
		},
		"returns a name-to-content map for the top-level files": {
			Files:    map[string]string{"a.puml": "@startuml\n", "b.puml": "@enduml\n"},
			Expected: map[string]string{"a.puml": "@startuml\n", "b.puml": "@enduml\n"},
		},
		"returns an empty file as an empty string": {
			Files:    map[string]string{"empty.puml": ""},
			Expected: map[string]string{"empty.puml": ""},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "target")
			if test.Files != nil {
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatalf("MkdirAll: %s", err)
				}
				for name, content := range test.Files {
					if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
						t.Fatalf("WriteFile: %s", err)
					}
				}
			}

			got, err := cli.ReadDirContents(dir)
			if err != nil {
				t.Fatalf("ReadDirContents(%q) should succeed, but got error: %s", dir, err)
			}
			if !maps.Equal(test.Expected, got) {
				t.Errorf("ReadDirContents(%q) mismatch (-want +got):\n%s", dir, cmp.Diff(test.Expected, got))
			}
		})
	}
}
