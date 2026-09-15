package cmd

import (
	"testing"

	"github.com/Kuniwak/puml-tools/cli"
)

func TestParseOptions(t *testing.T) {
	const (
		seqA = "testdata/seq_a.puml"
		seqB = "testdata/seq_b.puml"
	)

	tests := map[string]struct {
		Args      []string
		WantError bool
		Check     func(t *testing.T, o *Options)
	}{
		"multiple files accepted": {
			Args: []string{seqA, seqB},
			Check: func(t *testing.T, o *Options) {
				if len(o.SeqInputs) != 2 {
					t.Errorf("len(SeqInputs) = %d, want 2", len(o.SeqInputs))
				}
				if o.SeqInputs[0].Name != seqA || o.SeqInputs[1].Name != seqB {
					t.Errorf("names = %q, %q, want %q, %q", o.SeqInputs[0].Name, o.SeqInputs[1].Name, seqA, seqB)
				}
			},
		},
		"two stdin dashes error": {
			Args:      []string{"-", "-"},
			WantError: true,
		},
		"name flag": {
			Args: []string{"-name", "Ctx", seqA},
			Check: func(t *testing.T, o *Options) {
				if o.Name != "Ctx" {
					t.Errorf("Name = %q, want %q", o.Name, "Ctx")
				}
			},
		},
		"version flag": {
			Args: []string{"-v"},
			Check: func(t *testing.T, o *Options) {
				if !o.CommonOptions.Version {
					t.Error("want Version true")
				}
			},
		},
		"help flag": {
			Args: []string{"-h"},
			Check: func(t *testing.T, o *Options) {
				if !o.CommonOptions.Help {
					t.Error("want Help true")
				}
			},
		},
		"default stdin": {
			Args: []string{},
			Check: func(t *testing.T, o *Options) {
				if len(o.SeqInputs) != 1 {
					t.Errorf("len(SeqInputs) = %d, want 1 (stdin)", len(o.SeqInputs))
				}
				if o.SeqInputs[0].Name != "-" {
					t.Errorf("name = %q, want %q", o.SeqInputs[0].Name, "-")
				}
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			spy := cli.SpyProcInout("")
			o, err := ParseOptions(test.Args, spy.NewProcInout())
			if test.WantError {
				if err == nil {
					t.Fatal("want error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, in := range o.SeqInputs {
				defer in.Reader.Close()
			}
			if test.Check != nil {
				test.Check(t, o)
			}
		})
	}
}
