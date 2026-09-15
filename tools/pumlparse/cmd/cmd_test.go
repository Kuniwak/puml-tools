package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/tools/pumlparse/pumlpng"
	"github.com/google/go-cmp/cmp"
)

func writeTempPNG(t *testing.T, data []byte) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		t.Fatal(err)
	}
	return f.Name()
}

func TestMainCommandByArgs(t *testing.T) {
	tests := map[string]struct {
		Stdin          string
		StdinBytes     []byte
		Args           []string
		ArgsFunc       func(t *testing.T) []string
		ExpectedStatus int
		Expected       string
	}{
		"stdin plain text": {
			Stdin:          "@startuml\n@enduml\n",
			Args:           []string{},
			ExpectedStatus: 0,
			Expected:       "@startuml\n@enduml\n",
		},
		"dash arg reads stdin": {
			Stdin:          "hello\n",
			Args:           []string{"-"},
			ExpectedStatus: 0,
			Expected:       "hello\n",
		},
		"too many args": {
			Args:           []string{"a.png", "b.png"},
			ExpectedStatus: 1,
		},
		"file path plain text": {
			Args:           []string{"testdata/hello.puml"},
			ExpectedStatus: 0,
			Expected:       "@startuml\nhello\n@enduml\n",
		},
		"file path PNG with metadata": {
			ArgsFunc: func(t *testing.T) []string {
				return []string{writeTempPNG(t, pumlpng.NewPUMLPNG("@startuml\nhello\n@enduml\n"))}
			},
			ExpectedStatus: 0,
			Expected:       "@startuml\nhello\n@enduml\n",
		},
		"file path PNG no metadata": {
			ArgsFunc: func(t *testing.T) []string {
				return []string{writeTempPNG(t, pumlpng.NewPNG())}
			},
			ExpectedStatus: 1,
		},
		"stdin PNG with metadata": {
			StdinBytes:     pumlpng.NewPUMLPNG("@startuml\nhello\n@enduml\n"),
			Args:           []string{},
			ExpectedStatus: 0,
			Expected:       "@startuml\nhello\n@enduml\n",
		},
		"stdin PNG no metadata": {
			StdinBytes:     pumlpng.NewPNG(),
			Args:           []string{},
			ExpectedStatus: 1,
		},
		"default round-trips sequence diagram": {
			Stdin:          "@startuml\nparticipant A\nparticipant B\nA -> B: hello\ngroup g\n  B --> A: world\nend\n@enduml\n",
			Args:           []string{},
			ExpectedStatus: 0,
			Expected:       "@startuml\nparticipant A\nparticipant B\nA -> B: hello\ngroup g\n  B --> A: world\nend\n@enduml\n",
		},
		"raw flag passes through plain text": {
			Stdin:          "@startuml\nparticipant A\nparticipant B\nA -> B: hello\ngroup g\n  B --> A: world\nend\n@enduml\n",
			Args:           []string{"-raw"},
			ExpectedStatus: 0,
			Expected:       "@startuml\nparticipant A\nparticipant B\nA -> B: hello\ngroup g\n  B --> A: world\nend\n@enduml\n",
		},
		"raw flag with PNG metadata": {
			ArgsFunc: func(t *testing.T) []string {
				return []string{"-raw", writeTempPNG(t, pumlpng.NewPUMLPNG("@startuml\nhello\n@enduml\n"))}
			},
			ExpectedStatus: 0,
			Expected:       "@startuml\nhello\n@enduml\n",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			args := test.Args
			if test.ArgsFunc != nil {
				args = test.ArgsFunc(t)
			}

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

			exitStatus := MainCommandByArgs(args, spy.NewProcInout())
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
