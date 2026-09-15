package cmd

import (
	"bytes"
	"maps"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/tools/pumlparse/pumlpng"
	"github.com/Kuniwak/puml-tools/version"
	"github.com/google/go-cmp/cmp"
)

const inputSeq = "@startuml\nparticipant A\nparticipant B\nparticipant C\ngroup Login\n  A -> B: login\n  B -> A: ok\nend\ngroup Order / Confirm\n  B -> C: order\nend\n@enduml\n"

const cutLogin = "@startuml\nparticipant A\nparticipant B\ngroup Login\n  A -> B: login\n  B -> A: ok\nend\n@enduml\n"

const cutOrder = "@startuml\nparticipant B\nparticipant C\ngroup Order / Confirm\n  B -> C: order\nend\n@enduml\n"

func TestMainCommandByArgs(t *testing.T) {
	tests := map[string]struct {
		Stdin          string
		StdinBytes     []byte
		Args           []string
		ExpectedStatus int
		ExpectedStdout string
		ExpectedFiles  map[string]string
		ExpectedStderr string
	}{
		"diagram via file arg names files after the input": {
			Args:           []string{"-o", "%OUT%", "testdata/seq.puml"},
			ExpectedStatus: 0,
			ExpectedStdout: "%OUT%/seq_1_Login.puml\n%OUT%/seq_2_Order___Confirm.puml\n",
			ExpectedFiles: map[string]string{
				"seq_1_Login.puml":           cutLogin,
				"seq_2_Order___Confirm.puml": cutOrder,
			},
		},
		"diagram via stdin needs basename": {
			Stdin:          inputSeq,
			Args:           []string{"-o", "%OUT%"},
			ExpectedStatus: 1,
			ExpectedStderr: "-basename is required",
		},
		"basename flag names files": {
			Stdin:          inputSeq,
			Args:           []string{"-o", "%OUT%", "-basename", "flow"},
			ExpectedStatus: 0,
			ExpectedStdout: "%OUT%/flow_1_Login.puml\n%OUT%/flow_2_Order___Confirm.puml\n",
			ExpectedFiles: map[string]string{
				"flow_1_Login.puml":           cutLogin,
				"flow_2_Order___Confirm.puml": cutOrder,
			},
		},
		"basename flag overrides the input file name": {
			Args:           []string{"-o", "%OUT%", "-basename", "flow", "testdata/seq.puml"},
			ExpectedStatus: 0,
			ExpectedStdout: "%OUT%/flow_1_Login.puml\n%OUT%/flow_2_Order___Confirm.puml\n",
			ExpectedFiles: map[string]string{
				"flow_1_Login.puml":           cutLogin,
				"flow_2_Order___Confirm.puml": cutOrder,
			},
		},
		"a dash file arg means stdin": {
			Stdin:          inputSeq,
			Args:           []string{"-o", "%OUT%", "-basename", "flow", "-"},
			ExpectedStatus: 0,
			ExpectedStdout: "%OUT%/flow_1_Login.puml\n%OUT%/flow_2_Order___Confirm.puml\n",
			ExpectedFiles: map[string]string{
				"flow_1_Login.puml":           cutLogin,
				"flow_2_Order___Confirm.puml": cutOrder,
			},
		},
		"PNG diagram via stdin": {
			StdinBytes:     pumlpng.NewPUMLPNG(inputSeq),
			Args:           []string{"-o", "%OUT%", "-basename", "flow"},
			ExpectedStatus: 0,
			ExpectedStdout: "%OUT%/flow_1_Login.puml\n%OUT%/flow_2_Order___Confirm.puml\n",
			ExpectedFiles: map[string]string{
				"flow_1_Login.puml":           cutLogin,
				"flow_2_Order___Confirm.puml": cutOrder,
			},
		},
		"basename escaping the out dir is error": {
			Args:           []string{"-o", "%OUT%", "-basename", "../escaped", "testdata/seq.puml"},
			ExpectedStatus: 1,
			ExpectedStderr: `-basename "../escaped": basename must not contain a path separator`,
		},
		"missing out dir is error": {
			Args:           []string{"testdata/seq.puml"},
			ExpectedStatus: 1,
			ExpectedStderr: "-o is required",
		},
		"a message outside a group is error": {
			Stdin:          "@startuml\nparticipant A\nparticipant B\nA -> B: outside\ngroup G\n  A -> B\nend\n@enduml\n",
			Args:           []string{"-o", "%OUT%", "-basename", "flow"},
			ExpectedStatus: 1,
			ExpectedStderr: "message outside a group: line 4: A -> B: outside",
		},
		"no group is error": {
			Stdin:          "@startuml\nA -> B\n@enduml\n",
			Args:           []string{"-o", "%OUT%", "-basename", "flow"},
			ExpectedStatus: 1,
			ExpectedStderr: "no group found",
		},
		"too many file args": {
			Args:           []string{"-o", "%OUT%", "a.puml", "b.puml"},
			ExpectedStatus: 1,
			ExpectedStderr: "too many arguments",
		},
		"diagram open error": {
			Args:           []string{"-o", "%OUT%", "testdata/does-not-exist.puml"},
			ExpectedStatus: 1,
			ExpectedStderr: "testdata/does-not-exist.puml",
		},
		"help": {
			Args:           []string{"-h"},
			ExpectedStatus: 0,
			ExpectedStdout: "",
			ExpectedStderr: "Usage: pumlseqgroupsplit -o <out_dir> [options] (<file>|< <file>)",
		},
		"version": {
			Args:           []string{"-v"},
			ExpectedStatus: 0,
			ExpectedStdout: version.Version + "\n",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			outDir := filepath.Join(t.TempDir(), "out")
			args := make([]string, len(test.Args))
			for i, a := range test.Args {
				args[i] = strings.ReplaceAll(a, "%OUT%", outDir)
			}

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

			exitStatus := MainCommandByArgs(args, spy.NewProcInout())
			if exitStatus != test.ExpectedStatus {
				t.Errorf("MainCommandByArgs returned %d, want %d (stderr: %s)", exitStatus, test.ExpectedStatus, spy.Stderr.String())
			}

			expectedStdout := strings.ReplaceAll(test.ExpectedStdout, "%OUT%", outDir)
			if got := spy.Stdout.String(); got != expectedStdout {
				t.Errorf("stdout mismatch (-want +got):\n%s", cmp.Diff(expectedStdout, got))
			}

			written, err := cli.ReadDirContents(outDir)
			if err != nil {
				t.Fatalf("ReadDirContents: %s", err)
			}
			if !maps.Equal(test.ExpectedFiles, written) {
				t.Errorf("written files mismatch (-want +got):\n%s", cmp.Diff(test.ExpectedFiles, written))
			}

			if test.ExpectedStderr != "" {
				if got := spy.Stderr.String(); !strings.Contains(got, test.ExpectedStderr) {
					t.Errorf("stderr %q does not contain %q", got, test.ExpectedStderr)
				}
			}
		})
	}
}
