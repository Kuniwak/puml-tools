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

const inputSeq = "@startuml\nA -> B: login\nalt ok\n  B -> A: ok\nelse ng\n  B -> A: ng\nend\n@enduml\n"

const wideParSeq = "@startuml\npar\n  A -> B: a1\n  A -> B: a2\n  A -> B: a3\n  A -> B: a4\nelse\n  C -> D: b1\n  C -> D: b2\n  C -> D: b3\n  C -> D: b4\nelse\n  E -> F: c1\n  E -> F: c2\n  E -> F: c3\n  E -> F: c4\nend\n@enduml\n"

const warnParSeq = "@startuml\npar\n  A -> B: a1\n  A -> B: a2\n  A -> B: a3\n  A -> B: a4\n  A -> B: a5\nelse\n  C -> D: b1\n  C -> D: b2\n  C -> D: b3\n  C -> D: b4\n  C -> D: b5\nend\n@enduml\n"

func TestMainCommandByArgs(t *testing.T) {
	tests := map[string]struct {
		Stdin          string
		StdinBytes     []byte
		Args           []string
		ExpectedStatus int
		ExpectedStdout string
		ExpectedFiles  map[string]string
		ExpectedStderr string
		ExpectedQuiet  bool
	}{
		"diagram via file arg names files after the input": {
			Args:           []string{"-o", "%OUT%", "testdata/seq.puml"},
			ExpectedStatus: 0,
			ExpectedStdout: "%OUT%/seq_trace_1.tsv\n%OUT%/seq_trace_2.tsv\n",
			ExpectedFiles: map[string]string{
				"seq_trace_1.tsv": "login\nok\n",
				"seq_trace_2.tsv": "login\nng\n",
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
			ExpectedStdout: "%OUT%/flow_trace_1.tsv\n%OUT%/flow_trace_2.tsv\n",
			ExpectedFiles: map[string]string{
				"flow_trace_1.tsv": "login\nok\n",
				"flow_trace_2.tsv": "login\nng\n",
			},
		},
		"PNG diagram via stdin": {
			StdinBytes:     pumlpng.NewPUMLPNG(inputSeq),
			Args:           []string{"-o", "%OUT%", "-basename", "flow"},
			ExpectedStatus: 0,
			ExpectedStdout: "%OUT%/flow_trace_1.tsv\n%OUT%/flow_trace_2.tsv\n",
			ExpectedFiles: map[string]string{
				"flow_trace_1.tsv": "login\nok\n",
				"flow_trace_2.tsv": "login\nng\n",
			},
		},
		"loop unroll expands the loop up to the given turns": {
			Stdin:          "@startuml\nloop\n  A -> B: x\nend\n@enduml\n",
			Args:           []string{"-o", "%OUT%", "-basename", "flow", "-loop-unroll", "2"},
			ExpectedStatus: 0,
			ExpectedStdout: "%OUT%/flow_trace_1.tsv\n%OUT%/flow_trace_2.tsv\n",
			ExpectedFiles: map[string]string{
				"flow_trace_1.tsv": "x\n",
				"flow_trace_2.tsv": "x\nx\n",
			},
		},
		"break truncates the trace": {
			Stdin:          "@startuml\nA -> B: x\nbreak error\n  A -> B: e\nend\nA -> B: w\n@enduml\n",
			Args:           []string{"-o", "%OUT%", "-basename", "flow"},
			ExpectedStatus: 0,
			ExpectedStdout: "%OUT%/flow_trace_1.tsv\n",
			ExpectedFiles:  map[string]string{"flow_trace_1.tsv": "x\ne\n"},
		},
		"too many traces warns to use interleave sampling": {
			Stdin:          warnParSeq,
			Args:           []string{"-o", "%OUT%", "-basename", "flow"},
			ExpectedStatus: 0,
			ExpectedStderr: "-interleave-sampling",
		},
		"few traces do not warn": {
			Stdin:          inputSeq,
			Args:           []string{"-o", "%OUT%", "-basename", "flow"},
			ExpectedStatus: 0,
			ExpectedStdout: "%OUT%/flow_trace_1.tsv\n%OUT%/flow_trace_2.tsv\n",
			ExpectedQuiet:  true,
		},
		"interleave sampling thins out the traces deterministically": {
			Stdin:          wideParSeq,
			Args:           []string{"-o", "%OUT%", "-basename", "flow", "-interleave-sampling", "0.0001", "-seed", "1"},
			ExpectedStatus: 0,
			ExpectedFiles:  sampledFiles(t),
		},
		"interleave sampling of zero is error": {
			Stdin:          inputSeq,
			Args:           []string{"-o", "%OUT%", "-basename", "flow", "-interleave-sampling", "0"},
			ExpectedStatus: 1,
			ExpectedStderr: "-interleave-sampling must be in (0, 1]",
		},
		"interleave sampling out of range is error": {
			Stdin:          inputSeq,
			Args:           []string{"-o", "%OUT%", "-basename", "flow", "-interleave-sampling", "1.5"},
			ExpectedStatus: 1,
			ExpectedStderr: "-interleave-sampling must be in (0, 1]",
		},
		"negative loop unroll is error": {
			Stdin:          inputSeq,
			Args:           []string{"-o", "%OUT%", "-basename", "flow", "-loop-unroll", "-1"},
			ExpectedStatus: 1,
			ExpectedStderr: "-loop-unroll must not be negative",
		},
		"an unexpanded ref is error": {
			Stdin:          "@startuml\nref over A, B: other.puml\n@enduml\n",
			Args:           []string{"-o", "%OUT%", "-basename", "flow"},
			ExpectedStatus: 1,
			ExpectedStderr: "ref over",
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
			ExpectedStderr: "Usage: pumlseqtrace -o <out_dir> [options] (<file>|< <file>)",
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

			if test.ExpectedStdout != "" || test.ExpectedFiles != nil {
				if test.ExpectedStdout != "" {
					expectedStdout := strings.ReplaceAll(test.ExpectedStdout, "%OUT%", outDir)
					if got := spy.Stdout.String(); got != expectedStdout {
						t.Errorf("stdout mismatch (-want +got):\n%s", cmp.Diff(expectedStdout, got))
					}
				}
				written, err := cli.ReadDirContents(outDir)
				if err != nil {
					t.Fatalf("ReadDirContents: %s", err)
				}
				if test.ExpectedFiles != nil && !maps.Equal(test.ExpectedFiles, written) {
					t.Errorf("written files mismatch (-want +got):\n%s", cmp.Diff(test.ExpectedFiles, written))
				}
			}

			if test.ExpectedQuiet {
				if got := spy.Stderr.String(); got != "" {
					t.Errorf("stderr is %q, want empty", got)
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

func sampledFiles(t *testing.T) map[string]string {
	t.Helper()
	outDir := filepath.Join(t.TempDir(), "first")
	spy := &cli.ProcInoutSpy{
		Stdin:  bytes.NewReader([]byte(wideParSeq)),
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Env:    map[string]string{},
	}
	if status := MainCommandByArgs([]string{"-o", outDir, "-basename", "flow", "-interleave-sampling", "0.0001", "-seed", "1"}, spy.NewProcInout()); status != 0 {
		t.Fatalf("MainCommandByArgs returned %d (stderr: %s)", status, spy.Stderr.String())
	}
	files, err := cli.ReadDirContents(outDir)
	if err != nil {
		t.Fatalf("ReadDirContents: %s", err)
	}
	if len(files) == 0 {
		t.Fatalf("sampling produced no file; the test cannot tell determinism from emptiness")
	}
	return files
}
