package pumlseqexpand

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParsePath(t *testing.T) {
	tests := map[string]struct {
		Label     string
		Expected  string
		ExpectErr bool
	}{
		"relative path":                                  {Label: "./sub.puml", Expected: "./sub.puml"},
		"relative path without leading dot":              {Label: "sub.puml", Expected: "sub.puml"},
		"relative path through a parent directory":       {Label: "../other/sub.puml", Expected: "../other/sub.puml"},
		"trims surrounding whitespace":                   {Label: "  ./sub.puml  ", Expected: "./sub.puml"},
		"accepts a PNG as long as it is a relative path": {Label: "./sub.puml.png", Expected: "./sub.puml.png"},
		"rejects an empty label":                         {Label: "", ExpectErr: true},
		"rejects whitespace only":                        {Label: "   ", ExpectErr: true},
		"rejects an absolute path":                       {Label: "/abs/sub.puml", ExpectErr: true},
		"rejects a path rooted at the home directory":    {Label: "~/sub.puml", ExpectErr: true},
		"rejects a URL":                                  {Label: "https://example.com/sub.puml", ExpectErr: true},
		"rejects a bare scheme":                          {Label: "file://sub.puml", ExpectErr: true},
		"rejects a label with a volume name":             {Label: `C:\sub.puml`, ExpectErr: true},
		"rejects a non-path label with multiple words":   {Label: "Auth details", ExpectErr: true},
		"rejects multiple lines":                         {Label: "./a.puml\n./b.puml", ExpectErr: true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ParsePath(test.Label)
			if test.ExpectErr {
				if err == nil {
					t.Fatalf("ParsePath(%q) want error, got %q", test.Label, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParsePath(%q) unexpected error: %v", test.Label, err)
			}
			if got != test.Expected {
				t.Errorf("ParsePath mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}
