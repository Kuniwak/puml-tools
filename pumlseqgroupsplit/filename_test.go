package pumlseqgroupsplit

import (
	"testing"
)

func TestFileName(t *testing.T) {
	tests := map[string]struct {
		Basename string
		Idx      int
		Total    int
		Title    string
		Expected string
	}{
		"single digit total is not zero-padded":         {Basename: "seq", Idx: 1, Total: 9, Title: "Login", Expected: "seq_1_Login.puml"},
		"two-digit total pads the first digit":          {Basename: "seq", Idx: 1, Total: 10, Title: "Login", Expected: "seq_01_Login.puml"},
		"last index of a two-digit total is not padded": {Basename: "seq", Idx: 10, Total: 10, Title: "Login", Expected: "seq_10_Login.puml"},
		"three-digit total pads two digits":             {Basename: "seq", Idx: 1, Total: 100, Title: "Login", Expected: "seq_001_Login.puml"},
		"empty title is omitted":                        {Basename: "seq", Idx: 1, Total: 1, Title: "", Expected: "seq_1.puml"},
		"spaces in the title become underscores":        {Basename: "seq", Idx: 2, Total: 2, Title: "Order Confirm", Expected: "seq_2_Order_Confirm.puml"},
		"spaces in the basename are kept as-is":         {Basename: "my seq", Idx: 1, Total: 1, Title: "Login", Expected: "my seq_1_Login.puml"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := FileName(test.Basename, test.Idx, test.Total, test.Title)
			if got != test.Expected {
				t.Errorf("FileName(%q, %d, %d, %q) = %q, want %q", test.Basename, test.Idx, test.Total, test.Title, got, test.Expected)
			}
		})
	}
}

func TestSanitizeTitle(t *testing.T) {
	tests := map[string]struct {
		Title    string
		Expected string
	}{
		"safe characters are kept as-is":             {Title: "Login-1.v2", Expected: "Login-1.v2"},
		"empty string stays empty":                   {Title: "", Expected: ""},
		"half-width space becomes an underscore":     {Title: "Order Confirm", Expected: "Order_Confirm"},
		"full-width space becomes an underscore":     {Title: "Order　Confirm", Expected: "Order_Confirm"},
		"tab becomes an underscore":                  {Title: "a\tb", Expected: "a_b"},
		"slash becomes an underscore":                {Title: "Order/Confirm", Expected: "Order_Confirm"},
		"path separator sequence becomes underscore": {Title: "../a", Expected: ".._a"},
		"characters not allowed in filenames":        {Title: `a\b:c*d?e"f<g>h|i`, Expected: "a_b_c_d_e_f_g_h_i"},
		"non-ASCII characters are kept as-is":        {Title: "café", Expected: "café"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := SanitizeTitle(test.Title); got != test.Expected {
				t.Errorf("SanitizeTitle(%q) = %q, want %q", test.Title, got, test.Expected)
			}
		})
	}
}
