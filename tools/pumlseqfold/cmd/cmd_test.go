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

const inputSeqNamed = "@startuml Orig\nparticipant A\nparticipant B\nparticipant C\nA -> B: req\nB -> C: fwd\nC -> B: resp\nB -> A: resp\n@enduml\n"

const foldedSeq = "@startuml\nparticipant A\nparticipant \"Backend\" as BE\nA -> BE: req\nBE -> BE: fwd\nBE -> BE: resp\nBE -> A: resp\n@enduml\n"

const foldedSeqNamedFolded = "@startuml Folded\nparticipant A\nparticipant \"Backend\" as BE\nA -> BE: req\nBE -> BE: fwd\nBE -> BE: resp\nBE -> A: resp\n@enduml\n"

const foldedSeqNamedOrig = "@startuml Orig\nparticipant A\nparticipant \"Backend\" as BE\nA -> BE: req\nBE -> BE: fwd\nBE -> BE: resp\nBE -> A: resp\n@enduml\n"

const inputSeqEmptyBlock = "@startuml\nparticipant A\nparticipant B\nparticipant C\nA -> B: req\nopt leftover\nend\nB -> C: fwd\nC -> B: resp\nB -> A: resp\n@enduml\n"

func TestMainCommandByArgs(t *testing.T) {
	tests := map[string]struct {
		Stdin          string
		StdinBytes     []byte
		Args           []string
		ExpectedStatus int
		Expected       string
	}{
		"quoted target name containing a colon is still matched in note": {
			Stdin:          "@startuml\nparticipant \"A:1\"\nnote over \"A:1\" : hello\n@enduml\n",
			Args:           []string{"-ids", "A:1", "-new-id", "Z"},
			ExpectedStatus: 0,
			Expected:       "@startuml\nparticipant Z\nnote over Z : hello\n@enduml\n",
		},
		"missing ids flag is error": {
			Stdin:          inputSeq,
			Args:           []string{"-new-id", "BE"},
			ExpectedStatus: 1,
		},
		"missing new-id flag is error": {
			Stdin:          inputSeq,
			Args:           []string{"-ids", "B,C"},
			ExpectedStatus: 1,
		},
		"too many file args": {
			Args:           []string{"-ids", "B,C", "-new-id", "BE", "a.puml", "b.puml"},
			ExpectedStatus: 1,
		},
		"file open error": {
			Args:           []string{"-ids", "B,C", "-new-id", "BE", "testdata/does-not-exist.puml"},
			ExpectedStatus: 1,
		},
		"missing participant is error": {
			Stdin:          inputSeq,
			Args:           []string{"-ids", "X", "-new-id", "Y"},
			ExpectedStatus: 1,
		},
		"new-id collision is error": {
			Stdin:          inputSeq,
			Args:           []string{"-ids", "B", "-new-id", "A"},
			ExpectedStatus: 1,
		},
		"stdin worked example": {
			Stdin:          inputSeq,
			Args:           []string{"-ids", "B,C", "-new-id", "BE", "-new-desc", "Backend"},
			ExpectedStatus: 0,
			Expected:       foldedSeq,
		},
		"empty block pruned after fold": {
			Stdin:          inputSeqEmptyBlock,
			Args:           []string{"-ids", "B,C", "-new-id", "BE", "-new-desc", "Backend"},
			ExpectedStatus: 0,
			Expected:       foldedSeq,
		},
		"dash reads stdin": {
			Stdin:          inputSeq,
			Args:           []string{"-ids", "B,C", "-new-id", "BE", "-new-desc", "Backend", "-"},
			ExpectedStatus: 0,
			Expected:       foldedSeq,
		},
		"file arg worked example": {
			Args:           []string{"-ids", "B,C", "-new-id", "BE", "-new-desc", "Backend", "testdata/seq.puml"},
			ExpectedStatus: 0,
			Expected:       foldedSeq,
		},
		"empty desc produces bare declaration": {
			Stdin:          inputSeq,
			Args:           []string{"-ids", "B,C", "-new-id", "BE"},
			ExpectedStatus: 0,
			Expected:       "@startuml\nparticipant A\nparticipant BE\nA -> BE: req\nBE -> BE: fwd\nBE -> BE: resp\nBE -> A: resp\n@enduml\n",
		},
		"name flag is appended to startuml": {
			Stdin:          inputSeq,
			Args:           []string{"-ids", "B,C", "-new-id", "BE", "-new-desc", "Backend", "-name", "Folded"},
			ExpectedStatus: 0,
			Expected:       foldedSeqNamedFolded,
		},
		"omitted name preserves input name": {
			Stdin:          inputSeqNamed,
			Args:           []string{"-ids", "B,C", "-new-id", "BE", "-new-desc", "Backend"},
			ExpectedStatus: 0,
			Expected:       foldedSeqNamedOrig,
		},
		"name flag overwrites input name": {
			Stdin:          inputSeqNamed,
			Args:           []string{"-ids", "B,C", "-new-id", "BE", "-new-desc", "Backend", "-name", "Folded"},
			ExpectedStatus: 0,
			Expected:       foldedSeqNamedFolded,
		},
		"PNG via stdin": {
			StdinBytes:     pumlpng.NewPUMLPNG(inputSeq),
			Args:           []string{"-ids", "B,C", "-new-id", "BE", "-new-desc", "Backend"},
			ExpectedStatus: 0,
			Expected:       foldedSeq,
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
