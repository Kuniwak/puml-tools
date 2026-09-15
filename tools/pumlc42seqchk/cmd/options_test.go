package cmd

import (
	"slices"
	"testing"

	"github.com/Kuniwak/puml-tools/cli"
)

func TestParseOptions(t *testing.T) {
	const (
		seqSmall = "testdata/seq_small.puml"
		seqPart1 = "testdata/seq_part1.puml"
		seqPart2 = "testdata/seq_part2.puml"
		c4Eq     = "testdata/c4_eq.puml"
	)

	tests := map[string]struct {
		Args      []string
		WantError bool
		Check     func(t *testing.T, o *Options)
	}{
		"valid eq": {
			Args: []string{"-method", "eq", "-seq", seqSmall, "-c4", c4Eq},
			Check: func(t *testing.T, o *Options) {
				if o.Method != MethodEq {
					t.Errorf("Method = %q, want %q", o.Method, MethodEq)
				}
				if len(o.SeqInputs) != 1 || o.C4Reader == nil {
					t.Errorf("SeqInputs = %d, want 1 with c4 reader set", len(o.SeqInputs))
				}
			},
		},
		"default method is eq": {
			Args: []string{"-seq", seqSmall, "-c4", c4Eq},
			Check: func(t *testing.T, o *Options) {
				if o.Method != MethodEq {
					t.Errorf("Method = %q, want %q", o.Method, MethodEq)
				}
			},
		},
		"valid subset": {
			Args: []string{"-method", "subset", "-seq", seqSmall, "-c4", c4Eq},
			Check: func(t *testing.T, o *Options) {
				if o.Method != MethodSubset {
					t.Errorf("Method = %q, want %q", o.Method, MethodSubset)
				}
			},
		},
		"valid superset": {
			Args: []string{"-method", "superset", "-seq", seqSmall, "-c4", c4Eq},
			Check: func(t *testing.T, o *Options) {
				if o.Method != MethodSuperset {
					t.Errorf("Method = %q, want %q", o.Method, MethodSuperset)
				}
			},
		},
		"positional seq arguments": {
			Args: []string{"-method", "eq", "-c4", c4Eq, seqPart1, seqPart2},
			Check: func(t *testing.T, o *Options) {
				var names []string
				for _, in := range o.SeqInputs {
					names = append(names, in.Name)
				}
				want := []string{seqPart1, seqPart2}
				if !slices.Equal(want, names) {
					t.Errorf("SeqInputs names = %v, want %v", names, want)
				}
			},
		},
		"seq flag with positional arguments":   {Args: []string{"-seq", seqPart1, "-c4", c4Eq, seqPart2}, WantError: true},
		"two stdin among positional arguments": {Args: []string{"-c4", c4Eq, "-", "-"}, WantError: true},
		"invalid method":                       {Args: []string{"-method", "foo", "-seq", seqSmall, "-c4", c4Eq}, WantError: true},
		"missing seq":                          {Args: []string{"-c4", c4Eq}, WantError: true},
		"missing c4":                           {Args: []string{"-seq", seqSmall}, WantError: true},
		"both stdin":                           {Args: []string{"-seq", "-", "-c4", "-"}, WantError: true},
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
			if o.C4Reader != nil {
				defer o.C4Reader.Close()
			}
			if test.Check != nil {
				test.Check(t, o)
			}
		})
	}
}
