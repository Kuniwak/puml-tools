package cmd

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/pumlseqsubset"
	"github.com/Kuniwak/puml-tools/tools"
	"github.com/Kuniwak/puml-tools/tools/pumlparse/pumlpng"
	"github.com/Kuniwak/puml-tools/version"
	"github.com/google/go-cmp/cmp"
)

func TestMainCommandByOptionsRejectsUnknownMode(t *testing.T) {
	tests := map[string]struct {
		Mode pumlseqsubset.Mode
	}{
		"zero value":    {Mode: ""},
		"unknown value": {Mode: "unknown"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			options := &Options{
				CommonOptions: &tools.CommonOptions{},
				SuperReader:   io.NopCloser(strings.NewReader("@startuml\nA -> B: login\n@enduml\n")),
				SubReader:     io.NopCloser(strings.NewReader("@startuml\nA -> B: login\n@enduml\n")),
				SuperName:     "super.puml",
				SubName:       "sub.puml",
				Mode:          test.Mode,
			}

			spy := &cli.ProcInoutSpy{
				Stdin:  strings.NewReader(""),
				Stdout: &bytes.Buffer{},
				Stderr: &bytes.Buffer{},
				Env:    map[string]string{},
			}

			if _, err := MainCommandByOptions(options, spy.NewProcInout()); err == nil {
				t.Errorf("MainCommandByOptions returned no error for mode %q", test.Mode)
			}
		})
	}
}

func TestMainCommandByArgs(t *testing.T) {
	subMissing := "@startuml\nA -> B: login\nB -> X: extra\nB -> A: done\n@enduml\n"

	tests := map[string]struct {
		Stdin          string
		StdinBytes     []byte
		Args           []string
		ExpectedStatus int
		Expected       string
		ExpectedStderr string
	}{
		"super with unexpanded ref fails as a precondition violation": {
			Args:           []string{"testdata/with_ref.puml", "testdata/sub_ok.puml"},
			ExpectedStatus: 2,
			ExpectedStderr: "testdata/with_ref.puml:5: unexpanded ref found: ref over A, B; run it through pumlseqexpand first\n",
		},
		"sub with unexpanded ref fails as a precondition violation": {
			Args:           []string{"testdata/super.puml", "testdata/with_ref.puml"},
			ExpectedStatus: 2,
			ExpectedStderr: "testdata/with_ref.puml:5: unexpanded ref found: ref over A, B; run it through pumlseqexpand first\n",
		},
		"subset exits 0 with no output": {
			Args:           []string{"-mode", "allow-subseq", "testdata/super.puml", "testdata/sub_ok.puml"},
			ExpectedStatus: 0,
			Expected:       "",
		},
		"exact match exits 0": {
			Args:           []string{"-mode", "allow-subseq", "testdata/super.puml", "testdata/super.puml"},
			ExpectedStatus: 0,
			Expected:       "",
		},
		"missed messages exit 1 and list the offending lines": {
			Args:           []string{"-mode", "allow-subseq", "testdata/super.puml", "testdata/sub_missing.puml"},
			ExpectedStatus: 1,
			Expected:       "B\tX\n",
		},
		"reads sub from stdin": {
			Stdin:          subMissing,
			Args:           []string{"-mode", "allow-subseq", "testdata/super.puml", "-"},
			ExpectedStatus: 1,
			Expected:       "B\tX\n",
		},
		"reads sub from PNG metadata": {
			StdinBytes:     pumlpng.NewPUMLPNG(subMissing),
			Args:           []string{"-mode", "allow-subseq", "testdata/super.puml", "-"},
			ExpectedStatus: 1,
			Expected:       "B\tX\n",
		},
		"par branches are order-free, so reversed branch order still exits 0": {
			Args:           []string{"-mode", "allow-subseq", "testdata/par_super.puml", "testdata/par_sub.puml"},
			ExpectedStatus: 0,
			Expected:       "",
		},
		"a message missing from the par branches is detected as a residual": {
			Args:           []string{"-mode", "allow-subseq", "testdata/par_super.puml", "testdata/par_sub_extra.puml"},
			ExpectedStatus: 1,
			Expected:       "A\tZ\n",
		},
		"alt branches are mutually exclusive, so they are checked in written order": {
			Args:           []string{"-mode", "allow-subseq", "testdata/alt_super.puml", "testdata/par_sub.puml"},
			ExpectedStatus: 1,
			Expected:       "A\tC\n",
		},
		"the par on the sub side is also treated as order-free": {

			Args:           []string{"-mode", "allow-subseq", "testdata/alt_super.puml", "testdata/par_sub_par.puml"},
			ExpectedStatus: 0,
			Expected:       "",
		},
		"default is exact": {
			Args:           []string{"testdata/super.puml", "testdata/sub_ok.puml"},
			ExpectedStatus: 1,
			Expected:       "super\tB\tC\nsuper\tC\tB\n",
		},
		"exact reports leftovers on the super side as super rows too": {
			Args:           []string{"-mode", "exact", "testdata/super.puml", "testdata/sub_ok.puml"},
			ExpectedStatus: 1,
			Expected:       "super\tB\tC\nsuper\tC\tB\n",
		},
		"exact lists residuals from both sides, sub rows then super rows": {
			Args:           []string{"-mode", "exact", "testdata/super.puml", "testdata/sub_missing.puml"},
			ExpectedStatus: 1,
			Expected:       "sub\tB\tX\nsuper\tB\tC\nsuper\tC\tB\n",
		},
		"an exact match exits 0 under exact mode too": {
			Args:           []string{"-mode", "exact", "testdata/super.puml", "testdata/super.puml"},
			ExpectedStatus: 0,
			Expected:       "",
		},
		"par branches are order-free under exact mode too": {
			Args:           []string{"-mode", "exact", "testdata/par_super.puml", "testdata/par_sub.puml"},
			ExpectedStatus: 0,
			Expected:       "",
		},
		"an invalid -mode exits 2 and prints the reason to stderr": {
			Args:           []string{"-mode", "unknown", "testdata/super.puml", "testdata/sub_ok.puml"},
			ExpectedStatus: 2,
			Expected:       "",
			ExpectedStderr: "invalid -mode \"unknown\": must be \"exact\" or \"allow-subseq\"\n",
		},
		"a nonexistent file exits 2": {
			Args:           []string{"does-not-exist.puml", "testdata/sub_ok.puml"},
			ExpectedStatus: 2,
			Expected:       "",
		},
		"too few arguments exits 2": {
			Args:           []string{"testdata/super.puml"},
			ExpectedStatus: 2,
			Expected:       "",
		},
		"too many arguments exits 2": {
			Args:           []string{"a.puml", "b.puml", "c.puml"},
			ExpectedStatus: 2,
			Expected:       "",
		},
		"both arguments as stdin exits 2": {
			Args:           []string{"-", "-"},
			ExpectedStatus: 2,
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

			if test.ExpectedStderr != "" {
				if got := spy.Stderr.String(); got != test.ExpectedStderr {
					t.Errorf("stderr mismatch (-want +got):\n%s", cmp.Diff(test.ExpectedStderr, got))
				}
			}
		})
	}
}
