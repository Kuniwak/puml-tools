package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/tools/pumlparse/pumlpng"
	"github.com/Kuniwak/puml-tools/version"
	"github.com/google/go-cmp/cmp"
)

const wantTSV = "id\tkind\texternal\tboundary\tlabel\n" +
	"web\tSystem\tfalse\tshop\tWeb Storefront\n" +
	"order_svc\tSystem\tfalse\tbackend\tOrder Service\n" +
	"inventory\tSystem\tfalse\tbackend\tInventory DB\n"

func TestMainCommandByArgs(t *testing.T) {
	input, err := os.ReadFile("testdata/c4.puml")
	if err != nil {
		t.Fatal(err)
	}

	tests := map[string]struct {
		Stdin          string
		StdinBytes     []byte
		Args           []string
		ExpectedStatus int
		Expected       string
	}{
		"lists boundary members as tsv (stdin)": {
			Stdin:          string(input),
			Args:           []string{"-boundary", "shop"},
			ExpectedStatus: 0,
			Expected:       wantTSV,
		},
		"dash reads stdin": {
			Stdin:          string(input),
			Args:           []string{"-boundary", "shop", "-"},
			ExpectedStatus: 0,
			Expected:       wantTSV,
		},
		"file arg": {
			Args:           []string{"-boundary", "shop", "testdata/c4.puml"},
			ExpectedStatus: 0,
			Expected:       wantTSV,
		},
		"inner boundary": {
			Stdin:          string(input),
			Args:           []string{"-boundary", "backend"},
			ExpectedStatus: 0,
			Expected: "id\tkind\texternal\tboundary\tlabel\n" +
				"order_svc\tSystem\tfalse\tbackend\tOrder Service\n" +
				"inventory\tSystem\tfalse\tbackend\tInventory DB\n",
		},
		"PNG via stdin": {
			StdinBytes:     pumlpng.NewPUMLPNG(string(input)),
			Args:           []string{"-boundary", "shop"},
			ExpectedStatus: 0,
			Expected:       wantTSV,
		},
		"missing boundary flag is error": {
			Stdin:          string(input),
			Args:           []string{},
			ExpectedStatus: 1,
		},
		"boundary not found is error": {
			Stdin:          string(input),
			Args:           []string{"-boundary", "nope"},
			ExpectedStatus: 1,
		},
		"too many file args": {
			Args:           []string{"-boundary", "shop", "a.puml", "b.puml"},
			ExpectedStatus: 1,
		},
		"file open error": {
			Args:           []string{"-boundary", "shop", "testdata/does-not-exist.puml"},
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
