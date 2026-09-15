package outfiles

import (
	"maps"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/google/go-cmp/cmp"
)

func TestWriteFiles(t *testing.T) {
	tests := map[string]struct {
		Existing      map[string]string
		ExistingDirs  []string
		Files         []File
		ExpectedNames []string
		ExpectedErr   bool
		ExpectedFiles map[string]string
	}{
		"creates the output directory and writes everything when it does not exist": {
			Files:         []File{{Name: "a.puml", Content: "A\n"}, {Name: "b.puml", Content: "B\n"}},
			ExpectedNames: []string{"a.puml", "b.puml"},
			ExpectedFiles: map[string]string{"a.puml": "A\n", "b.puml": "B\n"},
		},
		"overwrites an existing file with the same name": {
			Existing:      map[string]string{"a.puml": "OLD\n"},
			Files:         []File{{Name: "a.puml", Content: "NEW\n"}},
			ExpectedNames: []string{"a.puml"},
			ExpectedFiles: map[string]string{"a.puml": "NEW\n"},
		},
		"leaves unrelated existing files alone": {
			Existing:      map[string]string{"keep.puml": "KEEP\n"},
			Files:         []File{{Name: "a.puml", Content: "A\n"}},
			ExpectedNames: []string{"a.puml"},
			ExpectedFiles: map[string]string{"keep.puml": "KEEP\n", "a.puml": "A\n"},
		},
		"writes nothing if even one destination cannot be written": {
			ExistingDirs:  []string{"b.puml"},
			Files:         []File{{Name: "a.puml", Content: "A\n"}, {Name: "b.puml", Content: "B\n"}},
			ExpectedErr:   true,
			ExpectedFiles: nil,
		},
		"writes nothing for an empty File slice": {
			Files:         nil,
			ExpectedNames: nil,
			ExpectedFiles: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "out")
			if test.Existing != nil || test.ExistingDirs != nil {
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatalf("MkdirAll: %s", err)
				}
			}
			for name, content := range test.Existing {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
					t.Fatalf("WriteFile: %s", err)
				}
			}
			for _, name := range test.ExistingDirs {
				if err := os.MkdirAll(filepath.Join(dir, name), 0o755); err != nil {
					t.Fatalf("MkdirAll: %s", err)
				}
			}

			paths, err := WriteFiles(dir, test.Files)
			if test.ExpectedErr {
				if err == nil {
					t.Fatalf("WriteFiles should fail, but got no error (paths: %v)", paths)
				}
				if len(paths) > 0 {
					t.Errorf("WriteFiles returned paths %v on failure, want none", paths)
				}
			} else {
				if err != nil {
					t.Fatalf("WriteFiles should succeed, but got error: %s", err)
				}
				var expected []string
				for _, name := range test.ExpectedNames {
					expected = append(expected, filepath.Join(dir, name))
				}
				if !slices.Equal(expected, paths) {
					t.Errorf("returned paths mismatch (-want +got):\n%s", cmp.Diff(expected, paths))
				}
			}

			written, err := cli.ReadDirContents(dir)
			if err != nil {
				t.Fatalf("ReadDirContents: %s", err)
			}
			if !maps.Equal(test.ExpectedFiles, written) {
				t.Errorf("written files mismatch (-want +got):\n%s", cmp.Diff(test.ExpectedFiles, written))
			}
		})
	}
}
