package cmd

import (
	"bytes"
	"testing"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/tools/pumlparse/pumlpng"
	"github.com/Kuniwak/puml-tools/version"
	"github.com/google/go-cmp/cmp"
)

const inputSeq = "@startuml\nparticipant A\nparticipant B\nA -> B: req\nA -> A: internal\nB -> B: log\nB -> A: resp\n@enduml\n"

const inputSeqNamed = "@startuml Orig\nparticipant A\nparticipant B\nA -> B: req\nA -> A: internal\nB -> B: log\nB -> A: resp\n@enduml\n"

const hiddenSeq = "@startuml\nparticipant A\nparticipant B\nA -> B: req\nB -> A: resp\n@enduml\n"

const hiddenSeqNamedHidden = "@startuml Hidden\nparticipant A\nparticipant B\nA -> B: req\nB -> A: resp\n@enduml\n"

const hiddenSeqNamedOrig = "@startuml Orig\nparticipant A\nparticipant B\nA -> B: req\nB -> A: resp\n@enduml\n"

const msgTSV = "A\tA\tinternal\nB\tB\tlog\n"

const inputSeqBlock = "@startuml\nparticipant A\nparticipant B\nA -> B: req\nopt detail\n  A -> A: internal\n  B -> B: log\nend\nB -> A: resp\n@enduml\n"

func TestMainCommandByArgs(t *testing.T) {
	tests := map[string]struct {
		Stdin          string
		StdinBytes     []byte
		Args           []string
		ExpectedStatus int
		Expected       string
	}{
		"msg-file path, diagram via stdin": {
			Stdin:          inputSeq,
			Args:           []string{"--msg-file", "testdata/hide.tsv"},
			ExpectedStatus: 0,
			Expected:       hiddenSeq,
		},
		"emptied block pruned after hide": {
			Stdin:          inputSeqBlock,
			Args:           []string{"--msg-file", "testdata/hide.tsv"},
			ExpectedStatus: 0,
			Expected:       hiddenSeq,
		},
		"msg-file path, diagram via file arg": {
			Args:           []string{"--msg-file", "testdata/hide.tsv", "testdata/selfmsg.puml"},
			ExpectedStatus: 0,
			Expected:       hiddenSeq,
		},
		"msg-file from stdin, diagram via file arg": {
			Stdin:          msgTSV,
			Args:           []string{"--msg-file", "-", "testdata/selfmsg.puml"},
			ExpectedStatus: 0,
			Expected:       hiddenSeq,
		},
		"PNG diagram via stdin, msg-file path": {
			StdinBytes:     pumlpng.NewPUMLPNG(inputSeq),
			Args:           []string{"--msg-file", "testdata/hide.tsv"},
			ExpectedStatus: 0,
			Expected:       hiddenSeq,
		},
		"name flag is appended to startuml": {
			Stdin:          inputSeq,
			Args:           []string{"--msg-file", "testdata/hide.tsv", "-name", "Hidden"},
			ExpectedStatus: 0,
			Expected:       hiddenSeqNamedHidden,
		},
		"omitted name preserves input name": {
			Stdin:          inputSeqNamed,
			Args:           []string{"--msg-file", "testdata/hide.tsv"},
			ExpectedStatus: 0,
			Expected:       hiddenSeqNamedOrig,
		},
		"name flag overwrites input name": {
			Stdin:          inputSeqNamed,
			Args:           []string{"--msg-file", "testdata/hide.tsv", "-name", "Hidden"},
			ExpectedStatus: 0,
			Expected:       hiddenSeqNamedHidden,
		},
		"missing msg-file is error": {
			Stdin:          inputSeq,
			Args:           []string{},
			ExpectedStatus: 1,
		},
		"both diagram and msg-file from stdin is error": {
			Stdin:          msgTSV,
			Args:           []string{"--msg-file", "-"},
			ExpectedStatus: 1,
		},
		"too many file args": {
			Args:           []string{"--msg-file", "testdata/hide.tsv", "a.puml", "b.puml"},
			ExpectedStatus: 1,
		},
		"msg-file open error": {
			Args:           []string{"--msg-file", "testdata/does-not-exist.tsv", "testdata/selfmsg.puml"},
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
