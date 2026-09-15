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

const foldedFull = "@startuml\n!include <C4/C4_Context>\n\n" +
	"Person(user, \"User\")\n" +
	"System_Ext(pay, \"Payment\")\n" +
	"System(shop, \"Shop\")\n\n" +
	"Rel(user, shop, \"browse\")\n" +
	"Rel(shop, pay, \"charge\")\n" +
	"@enduml\n"

const foldedFullNoDesc = "@startuml\n!include <C4/C4_Context>\n\n" +
	"Person(user, \"User\")\n" +
	"System_Ext(pay, \"Payment\")\n" +
	"System(shop, \"shop\")\n\n" +
	"Rel(user, shop, \"browse\")\n" +
	"Rel(shop, pay, \"charge\")\n" +
	"@enduml\n"

const foldedNamed = "@startuml Ctx\n!include <C4/C4_Context>\n\n" +
	"Person(user, \"User\")\n" +
	"System_Ext(pay, \"Payment\")\n" +
	"System(shop, \"Shop\")\n\n" +
	"Rel(user, shop, \"browse\")\n" +
	"Rel(shop, pay, \"charge\")\n" +
	"@enduml\n"

const foldedFilteredWeb = "@startuml\n!include <C4/C4_Container>\n\n" +
	"Person(user, \"User\")\n" +
	"System_Ext(pay, \"Payment\")\n" +
	"System(shop, \"Shop\")\n" +
	"Container(order_svc, \"Order Service\", \"Go\")\n\n" +
	"Rel(user, shop, \"browse\")\n" +
	"Rel(shop, order_svc, \"create\")\n" +
	"Rel(order_svc, pay, \"charge\")\n" +
	"@enduml\n"

func TestMainCommandByArgs(t *testing.T) {
	input, err := os.ReadFile("testdata/c4-lv2.puml")
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
		"ids happy (stdin)": {
			Stdin:          string(input),
			Args:           []string{"-ids", "web,order_svc", "-new-id", "shop", "-new-desc", "Shop"},
			ExpectedStatus: 0,
			Expected:       foldedFull,
		},
		"dash reads stdin": {
			Stdin:          string(input),
			Args:           []string{"-ids", "web,order_svc", "-new-id", "shop", "-new-desc", "Shop", "-"},
			ExpectedStatus: 0,
			Expected:       foldedFull,
		},
		"ids-tsv happy (file arg)": {
			Args:           []string{"-ids-tsv", "testdata/ids.tsv", "-new-id", "shop", "-new-desc", "Shop", "testdata/c4-lv2.puml"},
			ExpectedStatus: 0,
			Expected:       foldedFull,
		},
		"ids-tsv reordered columns": {
			Args:           []string{"-ids-tsv", "testdata/ids-reordered.tsv", "-new-id", "shop", "-new-desc", "Shop", "testdata/c4-lv2.puml"},
			ExpectedStatus: 0,
			Expected:       foldedFull,
		},
		"ids-tsv filtered subset": {
			Args:           []string{"-ids-tsv", "testdata/ids-filtered.tsv", "-new-id", "shop", "-new-desc", "Shop", "testdata/c4-lv2.puml"},
			ExpectedStatus: 0,
			Expected:       foldedFilteredWeb,
		},
		"empty new-desc uses new-id as label": {
			Stdin:          string(input),
			Args:           []string{"-ids", "web,order_svc", "-new-id", "shop"},
			ExpectedStatus: 0,
			Expected:       foldedFullNoDesc,
		},
		"name flag appended": {
			Stdin:          string(input),
			Args:           []string{"-ids", "web,order_svc", "-new-id", "shop", "-new-desc", "Shop", "-name", "Ctx"},
			ExpectedStatus: 0,
			Expected:       foldedNamed,
		},
		"PNG via stdin": {
			StdinBytes:     pumlpng.NewPUMLPNG(string(input)),
			Args:           []string{"-ids", "web,order_svc", "-new-id", "shop", "-new-desc", "Shop"},
			ExpectedStatus: 0,
			Expected:       foldedFull,
		},
		"both ids and ids-tsv is error": {
			Stdin:          string(input),
			Args:           []string{"-ids", "web", "-ids-tsv", "testdata/ids.tsv", "-new-id", "shop"},
			ExpectedStatus: 1,
		},
		"neither ids nor ids-tsv is error": {
			Stdin:          string(input),
			Args:           []string{"-new-id", "shop"},
			ExpectedStatus: 1,
		},
		"missing new-id is error": {
			Stdin:          string(input),
			Args:           []string{"-ids", "web,order_svc"},
			ExpectedStatus: 1,
		},
		"tsv without id column is error": {
			Args:           []string{"-ids-tsv", "testdata/ids-noheader.tsv", "-new-id", "shop", "testdata/c4-lv2.puml"},
			ExpectedStatus: 1,
		},
		"unknown id is error": {
			Stdin:          string(input),
			Args:           []string{"-ids", "web,nope", "-new-id", "shop"},
			ExpectedStatus: 1,
		},
		"ids-tsv file open error": {
			Args:           []string{"-ids-tsv", "testdata/does-not-exist.tsv", "-new-id", "shop", "testdata/c4-lv2.puml"},
			ExpectedStatus: 1,
		},
		"too many file args": {
			Args:           []string{"-ids", "web", "-new-id", "shop", "a.puml", "b.puml"},
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
