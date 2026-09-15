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

const seqFromStdin = "@startuml\nparticipant A\nparticipant B\nA -> B: req\nref over A, B : ./sub.puml\n@enduml\n"

func TestMainCommandByArgs(t *testing.T) {
	tests := map[string]struct {
		Stdin          string
		StdinBytes     []byte
		Args           []string
		ExpectedStatus int
		ExpectedStdout string
		StderrContains string
	}{
		"replaces a ref with the target's body": {
			Args:           []string{"testdata/seq.puml"},
			ExpectedStatus: 0,
			ExpectedStdout: "@startuml Seq\nparticipant A\nparticipant B\nA -> B: req\nA -> B: inner\nB -> A: ack\nB -> A: done\n@enduml\n",
		},
		"drops the target's diagram-wide directive and warns": {
			Args:           []string{"testdata/global.puml"},
			ExpectedStatus: 0,
			ExpectedStdout: "@startuml Global\nparticipant A\nparticipant B\nA -> B: req\nA -> B: inner\n@enduml\n",
			StderrContains: "testdata/global.puml:5: warning: testdata/global_sub.puml: dropped diagram-wide directive \"skinparam\" instead of inlining it",
		},
		"merges the target's box into the parent's box of the same name": {
			Args:           []string{"testdata/box.puml"},
			ExpectedStatus: 0,
			ExpectedStdout: "@startuml Box\nbox \"Ops\"\nactor \"A\" as a\nparticipant \"C\" as c\nend box\nparticipant \"B\" as b\na -> c: hello\nc -> b: hi\n@enduml\n",
			StderrContains: "testdata/box.puml:6: warning: testdata/box_sub.puml: dropped diagram-wide directive \"skinparam\"",
		},
		"fails when the parent and target boxes conflict": {
			Args:           []string{"testdata/box_conflict.puml"},
			ExpectedStatus: 2,
			StderrContains: "participant \"a\"'s box differs: outside any box vs box \"Ops\"",
		},
		"outputs unchanged when there is no ref": {
			Args:           []string{"testdata/noref.puml"},
			ExpectedStatus: 0,
			ExpectedStdout: "@startuml\nparticipant A\nparticipant B\nA -> B: plain\n@enduml\n",
		},
		"-name overrides the diagram name": {
			Args:           []string{"-name", "Expanded", "testdata/seq.puml"},
			ExpectedStatus: 0,
			ExpectedStdout: "@startuml Expanded\nparticipant A\nparticipant B\nA -> B: req\nA -> B: inner\nB -> A: ack\nB -> A: done\n@enduml\n",
		},
		"resolves stdin's relative paths from -base": {
			Stdin:          seqFromStdin,
			Args:           []string{"-base", "testdata"},
			ExpectedStatus: 0,
			ExpectedStdout: "@startuml\nparticipant A\nparticipant B\nA -> B: req\nA -> B: inner\nB -> A: ack\n@enduml\n",
		},
		"can expand a PNG input too": {
			StdinBytes:     pumlpng.NewPUMLPNG(seqFromStdin),
			Args:           []string{"-base", "testdata", "-"},
			ExpectedStatus: 0,
			ExpectedStdout: "@startuml\nparticipant A\nparticipant B\nA -> B: req\nA -> B: inner\nB -> A: ack\n@enduml\n",
		},
		"fails when a file argument and -base are given together": {
			Args:           []string{"-base", "testdata", "testdata/seq.puml"},
			ExpectedStatus: 2,
			StderrContains: "-base",
		},
		"fails when the label is not a relative path": {
			Args:           []string{"testdata/bad_label.puml"},
			ExpectedStatus: 2,
			StderrContains: "testdata/bad_label.puml:4: cannot read ref target",
		},
		"fails when the label is an absolute path": {
			Args:           []string{"testdata/abs_label.puml"},
			ExpectedStatus: 2,
			StderrContains: "testdata/abs_label.puml:4: ref label is an absolute path",
		},
		"fails on a scope violation": {
			Args:           []string{"testdata/scope.puml"},
			ExpectedStatus: 2,
			StderrContains: "testdata/scope.puml:4: testdata/leak.puml uses lifeline \"Z\", which is not covered by ref over A, B",
		},
		"fails on a reference cycle": {
			Args:           []string{"testdata/cycle_a.puml"},
			ExpectedStatus: 2,
			StderrContains: "testdata/cycle_a.puml:2: ",
		},
		"a nested ref's failure is reported with the target's name and line": {
			Args:           []string{"testdata/nested_bad.puml"},
			ExpectedStatus: 2,
			StderrContains: "testdata/nested_mid.puml:2: ref label is an absolute path",
		},
		"fails when the file does not exist": {
			Args:           []string{"testdata/does-not-exist.puml"},
			ExpectedStatus: 2,
			StderrContains: "testdata/does-not-exist.puml",
		},
		"fails when there are too many arguments": {
			Args:           []string{"a.puml", "b.puml"},
			ExpectedStatus: 2,
			StderrContains: "too many arguments",
		},
		"help": {
			Args:           []string{"-h"},
			ExpectedStatus: 0,
			StderrContains: "Usage: pumlseqexpand",
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

			status := MainCommandByArgs(test.Args, spy.NewProcInout())

			if status != test.ExpectedStatus {
				t.Errorf("status mismatch (-want +got):\n%s\nstderr: %s",
					cmp.Diff(test.ExpectedStatus, status), spy.Stderr.String())
			}

			if spy.Stdout.String() != test.ExpectedStdout {
				t.Errorf("stdout mismatch (-want +got):\n%s", cmp.Diff(test.ExpectedStdout, spy.Stdout.String()))
			}

			switch {
			case test.StderrContains == "" && spy.Stderr.String() != "":
				t.Errorf("stderr should be empty, but got %q", spy.Stderr.String())
			case !strings.Contains(spy.Stderr.String(), test.StderrContains):
				t.Errorf("stderr %q does not contain %q", spy.Stderr.String(), test.StderrContains)
			}
		})
	}
}
