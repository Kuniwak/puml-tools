package cmd

import (
	"slices"
	"testing"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/google/go-cmp/cmp"
)

func TestParseOptions(t *testing.T) {
	tests := map[string]struct {
		Args        []string
		WantErr     bool
		WantIDs     []string
		WantIDsTSV  bool
		WantNewID   string
		WantNewDesc string
		WantName    string
	}{
		"ids resolved from comma list": {
			Args:      []string{"-ids", "web,order_svc", "-new-id", "shop", "-new-desc", "Shop", "-name", "Ctx"},
			WantIDs:   []string{"web", "order_svc"},
			WantNewID: "shop", WantNewDesc: "Shop", WantName: "Ctx",
		},
		"ids-tsv opens a reader": {
			Args:       []string{"-ids-tsv", "testdata/ids.tsv", "-new-id", "shop", "testdata/c4-lv2.puml"},
			WantIDsTSV: true,
			WantNewID:  "shop",
		},
		"both ids and ids-tsv is error": {
			Args:    []string{"-ids", "web", "-ids-tsv", "testdata/ids.tsv", "-new-id", "shop"},
			WantErr: true,
		},
		"neither ids nor ids-tsv is error": {
			Args:    []string{"-new-id", "shop"},
			WantErr: true,
		},
		"empty ids list is error": {
			Args:    []string{"-ids", " , ", "-new-id", "shop"},
			WantErr: true,
		},
		"missing new-id is error": {
			Args:    []string{"-ids", "web"},
			WantErr: true,
		},
		"ids-tsv open error": {
			Args:    []string{"-ids-tsv", "testdata/does-not-exist.tsv", "-new-id", "shop"},
			WantErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			spy := cli.SpyProcInout("")
			opts, err := ParseOptions(test.Args, spy.NewProcInout())
			if test.WantErr {
				if err == nil {
					t.Fatalf("ParseOptions should error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseOptions returned error: %v", err)
			}
			if opts.Reader != nil {
				defer opts.Reader.Close()
			}
			if opts.IDsTSVReader != nil {
				defer opts.IDsTSVReader.Close()
			}
			if !slices.Equal(test.WantIDs, opts.IDs) {
				t.Errorf("IDs mismatch (-want +got):\n%s", cmp.Diff(test.WantIDs, opts.IDs))
			}
			if (opts.IDsTSVReader != nil) != test.WantIDsTSV {
				t.Errorf("IDsTSVReader present = %v, want %v", opts.IDsTSVReader != nil, test.WantIDsTSV)
			}
			if opts.NewID != test.WantNewID {
				t.Errorf("NewID = %q, want %q", opts.NewID, test.WantNewID)
			}
			if opts.NewDesc != test.WantNewDesc {
				t.Errorf("NewDesc = %q, want %q", opts.NewDesc, test.WantNewDesc)
			}
			if opts.Name != test.WantName {
				t.Errorf("Name = %q, want %q", opts.Name, test.WantName)
			}
		})
	}
}
