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
		"clean diagram exits zero": {
			Stdin:          "@startuml\na -> b\nb -> c\n@enduml\n",
			Args:           []string{},
			ExpectedStatus: 0,
			Expected:       "",
		},
		"break via stdin exits one": {
			Stdin:          "@startuml\na -> b\nc -> b\n@enduml\n",
			Args:           []string{},
			ExpectedStatus: 1,
			Expected:       "-\t3\tsender c has no preceding trigger\n",
		},
		"dash arg reads stdin": {
			Stdin:          "@startuml\na -> b\nc -> b\n@enduml\n",
			Args:           []string{"-"},
			ExpectedStatus: 1,
			Expected:       "-\t3\tsender c has no preceding trigger\n",
		},
		"file path with break": {
			Args:           []string{"testdata/broken.puml"},
			ExpectedStatus: 1,
			Expected:       "testdata/broken.puml\t3\tsender c has no preceding trigger\n",
		},
		"clean file exits zero": {
			Args:           []string{"testdata/clean.puml"},
			ExpectedStatus: 0,
			Expected:       "",
		},
		"ref without a preceding trigger exits one": {
			Args:           []string{"testdata/ref_orphan.puml"},
			ExpectedStatus: 1,
			Expected:       "testdata/ref_orphan.puml\t7\tref over c, d has no preceding trigger\n",
		},
		"ref establishes context so the following message is clean": {
			Args:           []string{"testdata/ref_clean.puml"},
			ExpectedStatus: 0,
			Expected:       "",
		},
		"stdin PNG with metadata": {
			StdinBytes:     pumlpng.NewPUMLPNG("@startuml\na -> b\nc -> b\n@enduml\n"),
			Args:           []string{},
			ExpectedStatus: 1,
			Expected:       "-\t3\tsender c has no preceding trigger\n",
		},
		"too many args": {
			Args:           []string{"a.puml", "b.puml"},
			ExpectedStatus: 1,
			Expected:       "",
		},
		"nonexistent file": {
			Args:           []string{"does-not-exist.puml"},
			ExpectedStatus: 1,
			Expected:       "",
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

			if got := spy.Stdout.String(); got != test.Expected {
				t.Errorf("stdout mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}
