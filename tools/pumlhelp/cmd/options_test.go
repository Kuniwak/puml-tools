package cmd

import (
	"testing"

	"github.com/Kuniwak/puml-tools/cli"
)

func TestParseOptions(t *testing.T) {
	tests := map[string]struct {
		Args      []string
		WantError bool
		Check     func(t *testing.T, o *Options)
	}{
		"no args selects all tools": {
			Args: []string{},
			Check: func(t *testing.T, o *Options) {
				if o.Short {
					t.Error("want Short false")
				}
				if len(o.ToolNames) != 0 {
					t.Errorf("len(ToolNames) = %d, want 0", len(o.ToolNames))
				}
			},
		},
		"short flag": {
			Args: []string{"-short"},
			Check: func(t *testing.T, o *Options) {
				if !o.Short {
					t.Error("want Short true")
				}
			},
		},
		"tool name stored": {
			Args: []string{"pumlparse", "pumlseqmsg"},
			Check: func(t *testing.T, o *Options) {
				if len(o.ToolNames) != 2 || o.ToolNames[0] != "pumlparse" || o.ToolNames[1] != "pumlseqmsg" {
					t.Errorf("ToolNames = %v, want [pumlparse pumlseqmsg]", o.ToolNames)
				}
			},
		},
		"self is a known tool": {
			Args: []string{"pumlhelp"},
			Check: func(t *testing.T, o *Options) {
				if len(o.ToolNames) != 1 || o.ToolNames[0] != "pumlhelp" {
					t.Errorf("ToolNames = %v, want [pumlhelp]", o.ToolNames)
				}
			},
		},
		"unknown tool error": {
			Args:      []string{"pumlrowidx"},
			WantError: true,
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
			if test.Check != nil {
				test.Check(t, o)
			}
		})
	}
}
