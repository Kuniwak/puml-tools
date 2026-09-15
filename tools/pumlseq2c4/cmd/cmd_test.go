package cmd

import (
	"os"
	"testing"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/version"
	"github.com/google/go-cmp/cmp"
)

func TestMainCommandByArgs(t *testing.T) {
	golden, err := os.ReadFile("testdata/seq_example.c4.golden")
	if err != nil {
		t.Fatal(err)
	}

	tests := map[string]struct {
		Stdin          string
		Args           []string
		ExpectedStatus int
		Expected       string
		ExpectedStderr string

		IgnoreStderr bool
	}{
		"stdin small diagram": {
			Stdin:          "@startuml\nactor a\nparticipant \"Bee\" as b\na -> b : hi\na -> a : self\n@enduml\n",
			Args:           []string{},
			ExpectedStatus: 0,
			Expected: "@startuml\n!include <C4/C4_Context>\n\n" +
				"Person(a, \"a\")\nSystem(b, \"Bee\")\n\n" +
				"Rel(a, b, \"hi\")\n@enduml\n",
		},
		"unquoted display name also becomes the label, not the alias": {
			Stdin:          "@startuml\nparticipant \"Order Service\" as order\nparticipant Cache as c\norder -> c : get\n@enduml\n",
			Args:           []string{},
			ExpectedStatus: 0,
			Expected: "@startuml\n!include <C4/C4_Context>\n\n" +
				"System(order, \"Order Service\")\nSystem(c, \"Cache\")\n\n" +
				"Rel(order, c, \"get\")\n@enduml\n",
		},
		"ref does not contribute to C4 (arrows in its body do not become Rel either)": {
			Stdin:          "@startuml\nparticipant a\nparticipant b\na -> b : real\nref over a, b\n  a -> b : inner\nend ref\n@enduml\n",
			Args:           []string{},
			ExpectedStatus: 0,
			Expected: "@startuml\n!include <C4/C4_Context>\n\n" +
				"System(a, \"a\")\nSystem(b, \"b\")\n\n" +
				"Rel(a, b, \"real\")\n@enduml\n",
			ExpectedStderr: "-:5: warning: ref over a, b does not contribute to C4 (pass the referenced diagram as well, or run it through pumlseqexpand)\n",
		},
		"warns that an unexpanded ref does not contribute to C4": {
			Args:           []string{"testdata/seq_ref.puml"},
			ExpectedStatus: 0,
			Expected: "@startuml\n!include <C4/C4_Context>\n\n" +
				"System(a, \"a\")\nSystem(b, \"b\")\n\n" +
				"Rel(a, b, \"real\")\n@enduml\n",
			ExpectedStderr: "testdata/seq_ref.puml:5: warning: ref over a, b does not contribute to C4 (pass the referenced diagram as well, or run it through pumlseqexpand)\n",
		},
		"file arg matches golden": {
			Args:           []string{"testdata/seq_example.puml"},
			ExpectedStatus: 0,
			Expected:       string(golden),
		},
		"two files union": {
			Args:           []string{"testdata/seq_a.puml", "testdata/seq_b.puml"},
			ExpectedStatus: 0,
			Expected: "@startuml\n!include <C4/C4_Context>\n\n" +
				"System(a, \"a\")\nSystem(b, \"b\")\nSystem(c, \"c\")\n\n" +
				"Rel(a, b, \"x\")\nRel(b, c, \"y\")\n@enduml\n",
		},
		"name flag sets diagram name": {
			Args:           []string{"-name", "Ctx", "testdata/seq_a.puml"},
			ExpectedStatus: 0,
			Expected: "@startuml Ctx\n!include <C4/C4_Context>\n\n" +
				"System(a, \"a\")\nSystem(b, \"b\")\n\n" +
				"Rel(a, b, \"x\")\n@enduml\n",
		},
		"boundary box makes a box become a boundary": {
			Stdin:          "@startuml\nactor user\nbox \"Order Service\"\n  participant tms\nend box\nuser -> tms : ask\n@enduml\n",
			Args:           []string{"-boundary", "box"},
			ExpectedStatus: 0,
			Expected: "@startuml\n!include <C4/C4_Context>\n\n" +
				"Person(user, \"user\")\n" +
				"System_Boundary(order_service, \"Order Service\") {\n" +
				"  System(tms, \"tms\")\n" +
				"}\n\n" +
				"Rel(user, tms, \"ask\")\n@enduml\n",
		},
		"a box does not become a boundary by default": {
			Stdin:          "@startuml\nactor user\nbox \"Order Service\"\n  participant tms\nend box\nuser -> tms : ask\n@enduml\n",
			Args:           []string{},
			ExpectedStatus: 0,
			Expected: "@startuml\n!include <C4/C4_Context>\n\n" +
				"Person(user, \"user\")\nSystem(tms, \"tms\")\n\n" +
				"Rel(user, tms, \"ask\")\n@enduml\n",
		},
		"error when the same id belongs to a different box across diagrams": {
			Args:           []string{"-boundary", "box", "testdata/seq_box_x.puml", "testdata/seq_box_y.puml"},
			ExpectedStatus: 1,
			Expected:       "",
			ExpectedStderr: "participant \"a\" belongs to different boxes across diagrams: \"X\" and \"Y\"\n",
		},
		"unknown -boundary value is an error": {
			Args:           []string{"-boundary", "nope"},
			ExpectedStatus: 1,
			Expected:       "",
			ExpectedStderr: "unknown -boundary value: \"nope\" (must be one of: none, box)\n",
		},
		"version": {
			Args:           []string{"-v"},
			ExpectedStatus: 0,
			Expected:       version.Version + "\n",
		},
		"help": {
			Args:           []string{"-h"},
			ExpectedStatus: 0,
			Expected:       "",
			IgnoreStderr:   true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			spy := cli.SpyProcInout(test.Stdin)
			status := MainCommandByArgs(test.Args, spy.NewProcInout())
			if status != test.ExpectedStatus {
				t.Errorf("status = %d, want %d (stderr: %s)", status, test.ExpectedStatus, spy.Stderr.String())
			}
			if got := spy.Stdout.String(); got != test.Expected {
				t.Errorf("stdout mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
			if got := spy.Stderr.String(); !test.IgnoreStderr && got != test.ExpectedStderr {
				t.Errorf("stderr mismatch (-want +got):\n%s", cmp.Diff(test.ExpectedStderr, got))
			}
		})
	}
}
