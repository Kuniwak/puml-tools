package cmd

import (
	"bytes"
	"testing"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/tools/pumlparse/pumlpng"
	"github.com/Kuniwak/puml-tools/version"
	"github.com/google/go-cmp/cmp"
)

const inputSeq = "@startuml\nparticipant A\nparticipant B\nparticipant C\nA -> B: req\nB -> C: fwd\nC -> B: resp\nB -> A: resp\n@enduml\n"

func TestMainCommandByArgs(t *testing.T) {
	tests := map[string]struct {
		Stdin          string
		StdinBytes     []byte
		Args           []string
		ExpectedStatus int
		Expected       string
	}{
		"common pair via file arg": {
			Args:           []string{"-common", "B,C", "testdata/seq.puml"},
			ExpectedStatus: 0,
			Expected:       "B\tC\tfwd\nC\tB\tresp\n",
		},
		"common pair via stdin": {
			Stdin:          inputSeq,
			Args:           []string{"-common", "B,C"},
			ExpectedStatus: 0,
			Expected:       "B\tC\tfwd\nC\tB\tresp\n",
		},
		"singleton yields self transitions and dedups": {
			Stdin:          "@startuml\nB -> B: loop\nA -> B: x\nB -> B: loop\n@enduml\n",
			Args:           []string{"-common", "B"},
			ExpectedStatus: 0,
			Expected:       "B\tB\tloop\n",
		},
		"no matching messages yields empty output": {
			Stdin:          inputSeq,
			Args:           []string{"-common", "A"},
			ExpectedStatus: 0,
			Expected:       "",
		},
		"PNG via stdin": {
			StdinBytes:     pumlpng.NewPUMLPNG(inputSeq),
			Args:           []string{"-common", "B,C"},
			ExpectedStatus: 0,
			Expected:       "B\tC\tfwd\nC\tB\tresp\n",
		},
		"missing common is error": {
			Args:           []string{"testdata/seq.puml"},
			ExpectedStatus: 1,
		},
		"too many file args": {
			Args:           []string{"-common", "B,C", "a.puml", "b.puml"},
			ExpectedStatus: 1,
		},
		"file open error": {
			Args:           []string{"-common", "B,C", "testdata/does-not-exist.puml"},
			ExpectedStatus: 1,
		},
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
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var stdin *bytes.Reader
			if test.StdinBytes != nil {
				stdin = bytes.NewReader(test.StdinBytes)
			} else {
				stdin = bytes.NewReader([]byte(test.Stdin))
			}
			spy := &cli.ProcInoutSpy{
				Stdin:  stdin,
				Stdout: &bytes.Buffer{},
				Stderr: &bytes.Buffer{},
				Env:    map[string]string{},
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
