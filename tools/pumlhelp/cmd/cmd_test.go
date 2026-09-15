package cmd

import (
	"strings"
	"testing"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/version"
	"github.com/google/go-cmp/cmp"
)

func TestMainCommandByArgs(t *testing.T) {
	tests := map[string]struct {
		Args           []string
		ExpectedStatus int
		Expected       string
	}{
		"help": {
			Args:           []string{"-h"},
			ExpectedStatus: 0,
			Expected:       "",
		},
		"version": {
			Args:           []string{"-v"},
			ExpectedStatus: 0,
			Expected:       version.Version + "\n",
		},
		"unknown tool is error": {
			Args:           []string{"pumlrowidx"},
			ExpectedStatus: 1,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			spy := cli.SpyProcInout("")
			exitStatus := MainCommandByArgs(test.Args, spy.NewProcInout())
			if exitStatus != test.ExpectedStatus {
				t.Errorf("MainCommandByArgs returned %d, want %d (stderr: %s)", exitStatus, test.ExpectedStatus, spy.Stderr.String())
			}

			if test.ExpectedStatus == 0 {
				if got := spy.Stdout.String(); got != test.Expected {
					t.Errorf("stdout mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
				}
			}
		})
	}
}

func TestMainCommandByArgsListsAllTools(t *testing.T) {
	spy := cli.SpyProcInout("")
	exitStatus := MainCommandByArgs(nil, spy.NewProcInout())
	if exitStatus != 0 {
		t.Fatalf("MainCommandByArgs returned %d (stderr: %s)", exitStatus, spy.Stderr.String())
	}

	got := spy.Stdout.String()
	for _, entry := range Registry {
		if entry.Hidden {
			if strings.Contains(got, "# "+entry.Name+"\n") {
				t.Errorf("stdout unexpectedly contains hidden tool heading %q", "# "+entry.Name)
			}
			continue
		}
		if !strings.Contains(got, "# "+entry.Name+"\n") {
			t.Errorf("stdout missing heading %q", "# "+entry.Name)
		}
	}
}

func TestMainCommandByArgsFiltersByName(t *testing.T) {
	spy := cli.SpyProcInout("")
	exitStatus := MainCommandByArgs([]string{"pumlparse"}, spy.NewProcInout())
	if exitStatus != 0 {
		t.Fatalf("MainCommandByArgs returned %d (stderr: %s)", exitStatus, spy.Stderr.String())
	}

	got := spy.Stdout.String()
	if !strings.Contains(got, "# pumlparse\n") {
		t.Errorf("stdout missing heading %q", "# pumlparse")
	}
	if strings.Contains(got, "# pumlseqgroupcut\n") {
		t.Errorf("stdout unexpectedly contains heading %q", "# pumlseqgroupcut")
	}
}

func TestMainCommandByArgsSelf(t *testing.T) {
	spy := cli.SpyProcInout("")
	exitStatus := MainCommandByArgs([]string{"pumlhelp"}, spy.NewProcInout())
	if exitStatus != 0 {
		t.Fatalf("MainCommandByArgs returned %d (stderr: %s)", exitStatus, spy.Stderr.String())
	}

	got := spy.Stdout.String()
	if !strings.Contains(got, "# pumlhelp\n") {
		t.Errorf("stdout missing heading %q", "# pumlhelp")
	}
}

func summaryOf(name string) string {
	for _, entry := range Registry {
		if entry.Name == name {
			return entry.Summary
		}
	}
	return ""
}

func TestMainCommandByArgsShort(t *testing.T) {
	t.Run("single tool is one exact line", func(t *testing.T) {
		spy := cli.SpyProcInout("")
		exitStatus := MainCommandByArgs([]string{"-short", "pumlparse"}, spy.NewProcInout())
		if exitStatus != 0 {
			t.Fatalf("MainCommandByArgs returned %d (stderr: %s)", exitStatus, spy.Stderr.String())
		}

		want := "pumlparse: " + summaryOf("pumlparse") + "\n"
		if got := spy.Stdout.String(); got != want {
			t.Errorf("stdout mismatch (-want +got):\n%s", cmp.Diff(want, got))
		}
	})

	t.Run("all tools listed as name: summary without headings", func(t *testing.T) {
		spy := cli.SpyProcInout("")
		exitStatus := MainCommandByArgs([]string{"-short"}, spy.NewProcInout())
		if exitStatus != 0 {
			t.Fatalf("MainCommandByArgs returned %d (stderr: %s)", exitStatus, spy.Stderr.String())
		}

		got := spy.Stdout.String()
		if strings.Contains(got, "# ") {
			t.Errorf("short output should not contain markdown headings:\n%s", got)
		}
		if strings.Contains(got, "```") {
			t.Errorf("short output should not contain code fences:\n%s", got)
		}
		if strings.Contains(got, "pumlhelp:") {
			t.Errorf("short output should not list hidden pumlhelp:\n%s", got)
		}
		for _, entry := range Registry {
			if entry.Hidden {
				continue
			}
			if !strings.Contains(got, entry.Name+": ") {
				t.Errorf("short output missing line for %q", entry.Name)
			}
		}
	})
}

func TestSelectEntries(t *testing.T) {
	registry := []ToolEntry{
		{Name: "pumlseqproj"},
		{Name: "pumlc4fold"},
		{Name: "pumlhelp", Hidden: true},
		{Name: "pumlparse"},
	}

	tests := map[string]struct {
		Names    []string
		Expected []string
	}{
		"empty names selects every visible tool in dictionary order": {
			Names:    nil,
			Expected: []string{"pumlc4fold", "pumlparse", "pumlseqproj"},
		},
		"names are honored in dictionary order, not in argument order": {
			Names:    []string{"pumlseqproj", "pumlc4fold", "pumlparse"},
			Expected: []string{"pumlc4fold", "pumlparse", "pumlseqproj"},
		},
		"hidden tool is selectable by name": {
			Names:    []string{"pumlhelp"},
			Expected: []string{"pumlhelp"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var got []string
			for _, entry := range SelectEntries(registry, test.Names) {
				got = append(got, entry.Name)
			}
			if diff := cmp.Diff(test.Expected, got); diff != "" {
				t.Errorf("selected tool order mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMainCommandByArgsSortsByName(t *testing.T) {
	tests := map[string]struct {
		Args     []string
		Expected []string
	}{
		"all tools are listed in dictionary order": {
			Args: []string{"-short"},
			Expected: []string{
				"pumlc42seqchk",
				"pumlc4bdyls",
				"pumlc4fold",
				"pumlparse",
				"pumlseq2c4",
				"pumlseqbreak",
				"pumlseqexpand",
				"pumlseqfold",
				"pumlseqgroupcut",
				"pumlseqgroupmask",
				"pumlseqgroupsplit",
				"pumlseqhide",
				"pumlseqmsg",
				"pumlseqproj",
				"pumlseqrefinement",
				"pumlseqsubset",
				"pumlseqtrace",
				"pumltodo",
			},
		},
		"explicit names are listed in dictionary order, not in argument order": {
			Args:     []string{"-short", "pumlseqgroupcut", "pumlc4fold", "pumlparse"},
			Expected: []string{"pumlc4fold", "pumlparse", "pumlseqgroupcut"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			spy := cli.SpyProcInout("")
			exitStatus := MainCommandByArgs(test.Args, spy.NewProcInout())
			if exitStatus != 0 {
				t.Fatalf("MainCommandByArgs returned %d (stderr: %s)", exitStatus, spy.Stderr.String())
			}

			var got []string
			for _, line := range strings.Split(strings.TrimSuffix(spy.Stdout.String(), "\n"), "\n") {
				got = append(got, strings.SplitN(line, ": ", 2)[0])
			}
			if diff := cmp.Diff(test.Expected, got); diff != "" {
				t.Errorf("tool order mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
