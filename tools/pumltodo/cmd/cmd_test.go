package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/tools/pumlparse/pumlpng"
	"github.com/Kuniwak/puml-tools/version"
	"github.com/google/go-cmp/cmp"
)

func TestMainCommandByArgs(t *testing.T) {
	tests := map[string]struct {
		Stdin          string
		StdinBytes     []byte
		Args           []string
		ExpectedStatus int
		Expected       string
	}{
		"stdin comment todo": {
			Stdin:          "'TODO: a\n",
			Args:           []string{},
			ExpectedStatus: 0,
			Expected:       "-\t1\tTODO: a\n",
		},
		"dash arg reads stdin": {
			Stdin:          "'TODO: d\n",
			Args:           []string{"-"},
			ExpectedStatus: 0,
			Expected:       "-\t1\tTODO: d\n",
		},
		"file path mixed sources": {
			Args:           []string{"testdata/mixed.puml"},
			ExpectedStatus: 0,
			Expected: "testdata/mixed.puml\t2\tTODO: first comment\n" +
				"testdata/mixed.puml\t4\tTODO implement reply\n" +
				"testdata/mixed.puml\t5\tTODO multi line note\n",
		},
		"no todos empty output": {
			Stdin:          "@startuml\n@enduml\n",
			Args:           []string{},
			ExpectedStatus: 0,
			Expected:       "",
		},
		"custom regexp matches fixme": {
			Stdin:          "'FIXME: x\n",
			Args:           []string{"--regexp", "(TODO|FIXME)"},
			ExpectedStatus: 0,
			Expected:       "-\t1\tFIXME: x\n",
		},
		"multiline note tsv quoting": {
			Stdin:          "note over A\nTODO alpha\nbeta line\nend note\n",
			Args:           []string{},
			ExpectedStatus: 0,
			Expected:       "-\t1\t\"TODO alpha\nbeta line\"\n",
		},
		"stdin PNG with metadata": {
			StdinBytes:     pumlpng.NewPUMLPNG("'TODO: in png\n"),
			Args:           []string{},
			ExpectedStatus: 0,
			Expected:       "-\t1\tTODO: in png\n",
		},
		"invalid regexp": {
			Args:           []string{"--regexp", "("},
			ExpectedStatus: 1,
		},
		"too many args": {
			Args:           []string{"a.puml", "b.puml"},
			ExpectedStatus: 1,
		},
		"nonexistent file": {
			Args:           []string{"does-not-exist.puml"},
			ExpectedStatus: 1,
		},
		"version short": {
			Args:           []string{"-v"},
			ExpectedStatus: 0,
			Expected:       version.Version + "\n",
		},
		"help": {
			Args:           []string{"-h"},
			ExpectedStatus: 0,
			Expected:       "",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var spy *cli.ProcInoutSpy
			if test.StdinBytes != nil {
				spy = &cli.ProcInoutSpy{
					Stdin:  bytes.NewReader(test.StdinBytes),
					Stdout: &bytes.Buffer{},
					Stderr: &bytes.Buffer{},
					Env:    map[string]string{},
				}
			} else {
				spy = &cli.ProcInoutSpy{
					Stdin:  strings.NewReader(test.Stdin),
					Stdout: &bytes.Buffer{},
					Stderr: &bytes.Buffer{},
					Env:    map[string]string{},
				}
			}

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
