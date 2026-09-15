package cmd

import (
	"testing"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/version"
	"github.com/google/go-cmp/cmp"
)

func TestMainCommandByArgs(t *testing.T) {
	const (
		seqSmall   = "testdata/seq_small.puml"
		c4Eq       = "testdata/c4_eq.puml"
		c4Subset   = "testdata/c4_subset.puml"
		c4Extra    = "testdata/c4_extra.puml"
		c4Superset = "testdata/c4_superset.puml"
		seqPart1   = "testdata/seq_part1.puml"
		seqPart2   = "testdata/seq_part2.puml"
		seqConf1   = "testdata/seq_conflict1.puml"
		seqConf2   = "testdata/seq_conflict2.puml"
		seqExmpl   = "testdata/seq_example.puml"
		c4Exmpl    = "testdata/c4_example.puml"
	)

	tests := map[string]struct {
		Args         []string
		WantStatus   int
		WantStdout   string
		WantNonEmpty bool
		WantStderr   string
	}{
		"eq aligned passes": {
			Args:       []string{"-method", "eq", "-seq", seqSmall, "-c4", c4Eq},
			WantStatus: 0,
			WantStdout: "OK: consistent\n",
		},
		"default method is eq": {
			Args:       []string{"-seq", seqSmall, "-c4", c4Eq},
			WantStatus: 0,
			WantStdout: "OK: consistent\n",
		},
		"subset aligned passes": {
			Args:       []string{"-method", "subset", "-seq", seqSmall, "-c4", c4Eq},
			WantStatus: 0,
			WantStdout: "OK: c4 ⊆ seq\n",
		},
		"subset strict passes": {
			Args:       []string{"-method", "subset", "-seq", seqSmall, "-c4", c4Subset},
			WantStatus: 0,
			WantStdout: "OK: c4 ⊆ seq\n",
		},
		"eq strict subset fails": {
			Args:       []string{"-method", "eq", "-seq", seqSmall, "-c4", c4Subset},
			WantStatus: 1,
			WantStdout: "participants only in SEQ (missing in C4): 1\n" +
				"  - c (\"c\")\n" +
				"messages only in SEQ (missing in C4): 1\n" +
				"  - b -> c\n",
		},
		"subset extra fails": {
			Args:       []string{"-method", "subset", "-seq", seqSmall, "-c4", c4Extra},
			WantStatus: 1,
			WantStdout: "elements only in C4 (not in SEQ): 1\n" +
				"  - z (\"z\")\n" +
				"relationships only in C4 (not in SEQ): 1\n" +
				"  - a -> z\n",
		},
		"superset aligned passes": {
			Args:       []string{"-method", "superset", "-seq", seqSmall, "-c4", c4Eq},
			WantStatus: 0,
			WantStdout: "OK: seq \u2286 c4\n",
		},
		"superset with extra c4 passes": {
			Args:       []string{"-method", "superset", "-seq", seqSmall, "-c4", c4Superset},
			WantStatus: 0,
			WantStdout: "OK: seq \u2286 c4\n",
		},
		"superset missing seq items fails": {
			Args:       []string{"-method", "superset", "-seq", seqSmall, "-c4", c4Subset},
			WantStatus: 1,
			WantStdout: "participants only in SEQ (not in C4): 1\n" +
				"  - c (\"c\")\n" +
				"messages only in SEQ (not in C4): 1\n" +
				"  - b -> c\n",
		},
		"real examples eq fails": {
			Args:         []string{"-method", "eq", "-seq", seqExmpl, "-c4", c4Exmpl},
			WantStatus:   1,
			WantNonEmpty: true,
		},
		"real examples subset fails": {
			Args:         []string{"-method", "subset", "-seq", seqExmpl, "-c4", c4Exmpl},
			WantStatus:   1,
			WantNonEmpty: true,
		},
		"eq over two seq files passes": {
			Args:       []string{"-method", "eq", "-c4", c4Eq, seqPart1, seqPart2},
			WantStatus: 0,
			WantStdout: "OK: consistent\n",
		},
		"eq over two seq files reports what only c4 has": {
			Args:       []string{"-method", "eq", "-c4", c4Superset, seqPart1, seqPart2},
			WantStatus: 1,
			WantStdout: "elements only in C4 (missing in SEQ): 1\n" +
				"  - z (\"z\")\n" +
				"relationships only in C4 (missing in SEQ): 1\n" +
				"  - a -> z\n",
		},
		"one seq file as positional arg passes": {
			Args:       []string{"-method", "eq", "-c4", c4Eq, seqSmall},
			WantStatus: 0,
			WantStdout: "OK: consistent\n",
		},
		"superset over two seq files passes": {
			Args:       []string{"-method", "superset", "-c4", c4Superset, seqPart1, seqPart2},
			WantStatus: 0,
			WantStdout: "OK: seq \u2286 c4\n",
		},
		"conflicting participant declarations error": {
			Args:       []string{"-method", "eq", "-c4", c4Eq, seqConf1, seqConf2},
			WantStatus: 2,
			WantStdout: "",
			WantStderr: "participant \"a\" conflicts between testdata/seq_conflict1.puml and testdata/seq_conflict2.puml: Person \"a\" vs System \"a\"\n",
		},
		"seq flag with positional args errors": {
			Args:       []string{"-seq", seqPart1, "-c4", c4Eq, seqPart2},
			WantStatus: 2,
			WantStdout: "",
		},
		"warns that an unexpanded ref does not contribute to C4 and continues checking": {
			Args:       []string{"-method", "eq", "-seq", "testdata/seq_ref.puml", "-c4", "testdata/c4_ref.puml"},
			WantStatus: 0,
			WantStdout: "OK: consistent\n",
			WantStderr: "testdata/seq_ref.puml:5: warning: ref over a, b does not contribute to C4 (pass the referenced diagram as well, or run it through pumlseqexpand)\n",
		},
		"version": {
			Args:       []string{"-v"},
			WantStatus: 0,
			WantStdout: version.Version + "\n",
		},
		"invalid method errors": {
			Args:       []string{"-method", "foo", "-seq", seqSmall, "-c4", c4Eq},
			WantStatus: 2,
			WantStdout: "",
		},
		"missing c4 errors": {
			Args:       []string{"-seq", seqSmall},
			WantStatus: 2,
			WantStdout: "",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			spy := cli.SpyProcInout("")
			status := MainCommandByArgs(test.Args, spy.NewProcInout())
			if status != test.WantStatus {
				t.Errorf("status = %d, want %d (stderr: %s)", status, test.WantStatus, spy.Stderr.String())
			}
			if test.WantNonEmpty {
				if spy.Stdout.Len() == 0 {
					t.Error("want non-empty stdout report")
				}
				return
			}
			if got := spy.Stdout.String(); got != test.WantStdout {
				t.Errorf("stdout mismatch (-want +got):\n%s", cmp.Diff(test.WantStdout, got))
			}
			if test.WantStderr != "" {
				if got := spy.Stderr.String(); got != test.WantStderr {
					t.Errorf("stderr mismatch (-want +got):\n%s", cmp.Diff(test.WantStderr, got))
				}
			}
		})
	}
}
