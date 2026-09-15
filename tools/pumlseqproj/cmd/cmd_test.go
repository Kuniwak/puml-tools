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

const inputSeq = "@startuml\nparticipant A\nparticipant B\nparticipant C\nA -> B: req\nA -> C: other\nB -> A: resp\n@enduml\n"

const inputSeqNamed = "@startuml Orig\nparticipant A\nparticipant B\nparticipant C\nA -> B: req\nA -> C: other\nB -> A: resp\n@enduml\n"

const projectedSeq = "@startuml\nparticipant A\nparticipant B\nA -> B: req\nB -> A: resp\n@enduml\n"

const projectedSeqNamedFocused = "@startuml Focused\nparticipant A\nparticipant B\nA -> B: req\nB -> A: resp\n@enduml\n"

const projectedSeqNamedOrig = "@startuml Orig\nparticipant A\nparticipant B\nA -> B: req\nB -> A: resp\n@enduml\n"

func TestMainCommandByArgs(t *testing.T) {
	tests := map[string]struct {
		Stdin          string
		StdinBytes     []byte
		Args           []string
		ExpectedStatus int
		Expected       string
		ExpectedStderr string
	}{
		"diagram via stdin": {
			Stdin:          inputSeq,
			Args:           []string{"-ids", "B"},
			ExpectedStatus: 0,
			Expected:       projectedSeq,
		},
		"diagram via file arg": {
			Args:           []string{"-ids", "B", "testdata/seq.puml"},
			ExpectedStatus: 0,
			Expected:       projectedSeq,
		},
		"a dash file arg means stdin": {
			Stdin:          inputSeq,
			Args:           []string{"-ids", "B", "-"},
			ExpectedStatus: 0,
			Expected:       projectedSeq,
		},
		"ids are comma separated": {
			Stdin:          inputSeq,
			Args:           []string{"-ids", "B,C"},
			ExpectedStatus: 0,
			Expected:       "@startuml\nparticipant A\nparticipant B\nparticipant C\nA -> B: req\nA -> C: other\nB -> A: resp\n@enduml\n",
		},
		"PNG diagram via stdin": {
			StdinBytes:     pumlpng.NewPUMLPNG(inputSeq),
			Args:           []string{"-ids", "B"},
			ExpectedStatus: 0,
			Expected:       projectedSeq,
		},
		"name flag is appended to startuml": {
			Stdin:          inputSeq,
			Args:           []string{"-ids", "B", "-name", "Focused"},
			ExpectedStatus: 0,
			Expected:       projectedSeqNamedFocused,
		},
		"omitted name preserves input name": {
			Stdin:          inputSeqNamed,
			Args:           []string{"-ids", "B"},
			ExpectedStatus: 0,
			Expected:       projectedSeqNamedOrig,
		},
		"name flag overwrites input name": {
			Stdin:          inputSeqNamed,
			Args:           []string{"-ids", "B", "-name", "Focused"},
			ExpectedStatus: 0,
			Expected:       projectedSeqNamedFocused,
		},
		"unknown id is error": {
			Stdin:          inputSeq,
			Args:           []string{"-ids", "Nobody"},
			ExpectedStatus: 1,
			ExpectedStderr: "participant not found: Nobody",
		},
		"missing ids is error": {
			Stdin:          inputSeq,
			Args:           []string{},
			ExpectedStatus: 1,
			ExpectedStderr: "-ids is required",
		},
		"too many file args": {
			Args:           []string{"-ids", "B", "a.puml", "b.puml"},
			ExpectedStatus: 1,
			ExpectedStderr: "too many arguments",
		},
		"diagram open error": {
			Args:           []string{"-ids", "B", "testdata/does-not-exist.puml"},
			ExpectedStatus: 1,
			ExpectedStderr: "testdata/does-not-exist.puml",
		},
		"help": {
			Args:           []string{"-h"},
			ExpectedStatus: 0,
			Expected:       "",
			ExpectedStderr: "Usage: pumlseqproj -ids <a,b,...> [options] (<file>|< <file>)",
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

			if test.ExpectedStderr != "" {
				if got := spy.Stderr.String(); !strings.Contains(got, test.ExpectedStderr) {
					t.Errorf("stderr %q does not contain %q", got, test.ExpectedStderr)
				}
			}
		})
	}
}
