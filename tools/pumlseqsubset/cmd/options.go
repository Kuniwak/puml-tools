package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/pumlseqsubset"
	"github.com/Kuniwak/puml-tools/tools"
)

type Options struct {
	CommonOptions *tools.CommonOptions
	SuperReader   io.ReadCloser
	SubReader     io.ReadCloser
	SuperName     string
	SubName       string
	Mode          pumlseqsubset.Mode
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("pumlseqsubset", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		fmt.Fprintf(inout.Stderr, `Usage: pumlseqsubset [options] <super> <sub>

Treats each sequence diagram as a sequence of (sender, receiver) pairs and
checks the relationship between <super> and <sub>. Message text and arrow
style are ignored. Exits 0 if the relationship holds, 1 if it doesn't, and
on failure lists the unmatched messages as TSV; the columns depend on -mode.
-mode exact (default) requires both diagrams to represent the same message
sequence, and reports the leftovers from either side as 3 columns:
(sub|super<TAB>sender<TAB>receiver). The first column says which diagram the
message appeared only in; sub's leftovers are listed in appearance order,
followed by super's.
-mode allow-subseq checks whether <sub> is a subsequence of <super> (i.e.
whether <super> can be obtained by inserting elements into <sub>), and
allows extra messages on the <super> side. (This mode is meant for cases
where <super> is a completed diagram and <sub> is the original, incomplete
one.) Any leftovers are necessarily on the <sub> side, so they're listed in
appearance order as 2 columns: (sender<TAB>receiver).
Either argument may be - to read from stdin.
Model: this is a structural comparison over one flat message sequence, not
an evaluation of UML's interaction semantics. Only par is given a meaning:
its branches run concurrently, so any linearization of them on either side
is allowed. Every other block (alt, opt, loop, group, critical, break) is
flattened by concatenating its branches in written order -- an alt is not
treated as a choice between branches, and a loop body is taken exactly
once. If you need alt as a choice and loop as repetition, use
pumlseqrefinement, which evaluates them as a CSP process instead.
Note that the number of states to check grows exponentially with the number
of par branches, multiplying several-fold per additional branch. Diagrams
with up to about 8 branches finish instantly, but a single par with 10 or
more branches can make the runtime impractical.
ref over blocks can't be represented as a message sequence, so an
unexpanded ref causes the command to exit with status 2 rather than being
silently ignored. Run it through pumlseqexpand first.
For PNG input, the embedded PlantUML metadata is read. Errors exit with
status 2.

Options:
`)
		flags.PrintDefaults()
		fmt.Fprintf(inout.Stderr, `
Examples:
  $ pumlseqsubset super.puml sub.puml; echo $?
  sub	B	X
  super	B	C
  super	C	B
  1

  $ pumlseqsubset -mode allow-subseq super.puml sub.puml; echo $?
  B	X
  1
`)
	}

	var mode string
	flags.StringVar(&mode, "mode", string(pumlseqsubset.ModeExact), "comparison mode exact|allow-subseq (default exact)")

	commonRawOptions := &tools.CommonRawOptions{}
	tools.DeclareCommonFlags(flags, commonRawOptions)

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return &Options{CommonOptions: &tools.CommonOptions{Help: true}}, nil
		}
		return nil, err
	}

	commonOptions, err := tools.ValidateCommonOptions(commonRawOptions)
	if err != nil {
		return nil, err
	}
	if commonOptions.Version {
		return &Options{CommonOptions: &commonOptions}, nil
	}

	parsedMode, err := pumlseqsubset.ParseMode(mode)
	if err != nil {
		return nil, fmt.Errorf("invalid -mode %q: must be %q or %q", mode, pumlseqsubset.ModeExact, pumlseqsubset.ModeAllowSubseq)
	}

	remaining := flags.Args()
	if len(remaining) != 2 {
		return nil, fmt.Errorf("requires exactly 2 file paths: <super> <sub>")
	}
	superName, subName := remaining[0], remaining[1]
	if superName == "-" && subName == "-" {
		return nil, errors.New("at most one argument can be - (stdin)")
	}

	superReader, err := cli.OpenInput(superName, inout)
	if err != nil {
		return nil, err
	}
	subReader, err := cli.OpenInput(subName, inout)
	if err != nil {
		superReader.Close()
		return nil, err
	}

	return &Options{
		CommonOptions: &commonOptions,
		SuperReader:   superReader,
		SubReader:     subReader,
		SuperName:     superName,
		SubName:       subName,
		Mode:          parsedMode,
	}, nil
}
