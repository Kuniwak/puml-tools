package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/version"
	"github.com/google/go-cmp/cmp"
)

const inputSeq = "@startuml\nparticipant A\nparticipant B\nparticipant C\nA -> B: login\ngroup Auth\n  B -> C: verify\n  C -> B: ok\nend\nB -> A: done\n@enduml\n"

const cutAuth = "@startuml\nparticipant B\nparticipant C\ngroup Auth\n  B -> C: verify\n  C -> B: ok\nend\n@enduml\n"

func TestMainCommandByArgs(t *testing.T) {
	tests := map[string]struct {
		Stdin          string
		Args           []string
		ExpectedStatus int
		Expected       string
	}{
		"missing exact flag is error": {
			Stdin:          inputSeq,
			Args:           []string{},
			ExpectedStatus: 1,
		},
		"too many file args": {
			Args:           []string{"-exact", "Auth", "a.puml", "b.puml"},
			ExpectedStatus: 1,
		},
		"file open error": {
			Args:           []string{"-exact", "Auth", "testdata/does-not-exist.puml"},
			ExpectedStatus: 1,
		},
		"zero matches is error": {
			Stdin:          inputSeq,
			Args:           []string{"-exact", "Nope"},
			ExpectedStatus: 1,
		},
		"stdin worked example": {
			Stdin:          inputSeq,
			Args:           []string{"-exact", "Auth"},
			ExpectedStatus: 0,
			Expected:       cutAuth,
		},
		"dash reads stdin": {
			Stdin:          inputSeq,
			Args:           []string{"-exact", "Auth", "-"},
			ExpectedStatus: 0,
			Expected:       cutAuth,
		},
		"file arg worked example": {
			Args:           []string{"-exact", "Auth", "testdata/auth.puml"},
			ExpectedStatus: 0,
			Expected:       cutAuth,
		},
		"name flag appended to startuml": {
			Stdin:          inputSeq,
			Args:           []string{"-exact", "Auth", "-name", "AuthExcerpt"},
			ExpectedStatus: 0,
			Expected:       "@startuml AuthExcerpt\n" + cutAuth[len("@startuml\n"):],
		},
		"nested group: outer match keeps inner": {
			Args:           []string{"-exact", "Outer", "testdata/nested.puml"},
			ExpectedStatus: 0,
			Expected:       "@startuml\nparticipant A\nparticipant B\nparticipant C\ngroup Outer\n  A -> B\n  group Inner\n    B -> C\n  end\nend\n@enduml\n",
		},
		"nested group: inner match": {
			Args:           []string{"-exact", "Inner", "testdata/nested.puml"},
			ExpectedStatus: 0,
			Expected:       "@startuml\nparticipant B\nparticipant C\n  group Inner\n    B -> C\n  end\n@enduml\n",
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
			spy := &cli.ProcInoutSpy{
				Stdin:  strings.NewReader(test.Stdin),
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
