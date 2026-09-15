package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/tools/pumlparse/pumlpng"
	"github.com/Kuniwak/puml-tools/version"
)

const detailInternalFromStdin = "@startuml\nparticipant A\nparticipant B\nparticipant C\nA -> B: req\nB -> C: secret\nC -> A: ack\n@enduml\n"

const detailFromStdin = "@startuml\nparticipant A\nparticipant B\nparticipant C\nA -> B: req\nC -> A: ack\n@enduml\n"

func TestMainCommandByArgs(t *testing.T) {
	tests := map[string]struct {
		Stdin          string
		StdinBytes     []byte
		Args           []string
		ExpectedStatus int
		ExpectedStdout string
		StderrContains string
	}{
		"refinement holds with map": {
			Args:           []string{"-map", "testdata/map.tsv", "testdata/abstract.puml", "testdata/detail.puml"},
			ExpectedStatus: 0,
		},
		"not a refinement when internal message stays visible": {
			Args:           []string{"-map", "testdata/map.tsv", "testdata/abstract.puml", "testdata/detail-internal.puml"},
			ExpectedStatus: 1,
			StderrContains: "Refinement does not hold",
		},
		"a detail containing an unexpanded ref aborts as a precondition violation": {
			Args:           []string{"-map", "testdata/map.tsv", "testdata/abstract.puml", "testdata/detail-ref.puml"},
			ExpectedStatus: 2,
			StderrContains: "testdata/detail-ref.puml:6: unexpanded ref found: ref over B, C; run it through pumlseqexpand first",
		},
		"an abstract containing an unexpanded ref aborts as a precondition violation": {
			Args:           []string{"testdata/detail-ref.puml", "testdata/detail-ref.puml"},
			ExpectedStatus: 2,
			StderrContains: "unexpanded ref found",
		},
		"identity refinement without map": {
			Args:           []string{"testdata/abstract.puml", "testdata/abstract.puml"},
			ExpectedStatus: 0,
		},
		"map image not in abstract is error": {
			Args:           []string{"-map", "testdata/map-bad.tsv", "testdata/abstract.puml", "testdata/detail.puml"},
			ExpectedStatus: 2,
		},
		"map row without abstract id is error": {
			Args:           []string{"-map", "testdata/map-invalid-row.tsv", "testdata/abstract.puml", "testdata/detail.puml"},
			ExpectedStatus: 2,
			StderrContains: "invalid map row",
		},
		"map and diagram both from stdin is error": {
			Stdin:          detailFromStdin,
			Args:           []string{"-map", "-", "testdata/abstract.puml", "-"},
			ExpectedStatus: 2,
			StderrContains: "cannot read both a diagram and -map from stdin",
		},

		"PNG detail from stdin via dash": {
			StdinBytes:     pumlpng.NewPUMLPNG(detailInternalFromStdin),
			Args:           []string{"-map", "testdata/map.tsv", "testdata/abstract.puml", "-"},
			ExpectedStatus: 1,
			StderrContains: "BE -> BE: secret",
		},
		"detail from stdin via dash": {
			Stdin:          detailFromStdin,
			Args:           []string{"-map", "testdata/map.tsv", "testdata/abstract.puml", "-"},
			ExpectedStatus: 0,
		},
		"missing file is error": {
			Args:           []string{"testdata/abstract.puml", "testdata/does-not-exist.puml"},
			ExpectedStatus: 2,
		},
		"one positional arg is error": {
			Args:           []string{"testdata/abstract.puml"},
			ExpectedStatus: 2,
		},
		"three positional args is error": {
			Args:           []string{"a.puml", "b.puml", "c.puml"},
			ExpectedStatus: 2,
		},
		"both stdin is error": {
			Args:           []string{"-", "-"},
			ExpectedStatus: 2,
		},
		"help": {
			Args:           []string{"-h"},
			ExpectedStatus: 0,
		},
		"version": {
			Args:           []string{"-v"},
			ExpectedStatus: 0,
			ExpectedStdout: version.Version + "\n",
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
			if test.ExpectedStdout != "" && spy.Stdout.String() != test.ExpectedStdout {
				t.Errorf("stdout = %q, want %q", spy.Stdout.String(), test.ExpectedStdout)
			}
			if test.StderrContains != "" && !strings.Contains(spy.Stderr.String(), test.StderrContains) {
				t.Errorf("stderr %q does not contain %q", spy.Stderr.String(), test.StderrContains)
			}
		})
	}
}
