package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/tools"
)

type Options struct {
	CommonOptions *tools.CommonOptions
	Base          string
	Name          string
	Input         tools.NamedInput
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("pumlseqexpand", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		fmt.Fprintf(inout.Stderr, `Usage: pumlseqexpand [options] (<file>|< <file>)

Replaces each ref over in a sequence diagram with the body of the diagram it
references (recursively). A ref's label must be a path relative to the
referencing file. Everything below is an error (exit code 2): a label that
isn't a relative path, a target that can't be read, a reference cycle, a
scope violation (the target uses a lifeline missing from the ref's over
clause), a participant declaration that conflicts across diagrams, and a
target that isn't a single diagram enclosed by @startuml and @enduml. Run it
over every diagram in CI to catch mistakes in how refs are written (find's
-exec doesn't propagate exit codes, so pipe through xargs as in the example
below).

A ref target must be a single diagram enclosed by @startuml and @enduml. Any
diagram-wide directive the target carries for rendering on its own
(title/caption/header/footer/legend/autonumber/skinparam/skin/scale/hide/
show/newpage) is dropped instead of inlined, and each drop is reported to
stderr as a warning (the parent diagram's own directive still applies, so its
appearance is unchanged). A directive written across multiple lines
(title/header/footer/legend) is dropped along with its whole body, up to
"end <keyword>". Comments, blank lines, and dividers (== x ==) are positional
and are always inlined as-is; the contents of a block comment are left alone
since they aren't a diagram directive.

A target's participant declarations are not left at the expansion site;
only the ones not already declared are appended to the end of the parent
diagram's declaration block. A declaration that was inside a box in the
target goes into the parent's box of the same name if one exists, otherwise
the whole box is appended to the end of the declaration block. Declarations
must agree across diagrams: declaring the same id with a different kind or
label is an error, and so is disagreeing about which box it belongs to (or
whether it's outside any box at all). A box models which lifelines belong
together, so diagrams joined by ref must place the same participant in the
same box. Errors and warnings are formatted as
"<referencing file>:<ref's line>: ...", so it's clear which ref is at fault
even when processing many diagrams at once.

When given a file argument, relative paths are resolved from that file's
directory (combining this with -base is an error). When reading from stdin,
use -base to set that starting directory (it defaults to the current
directory). A ref's label can't contain whitespace, so files with whitespace
in their names can't be referenced. The output's @startuml keeps the input's
name; pass -name to override it. Note: inlined bodies and hoisted
declarations carry over the target's own line terminators as-is, so the
output will mix CRLF and LF if the parent and target use different ones
(this doesn't affect how PlantUML interprets it).

Options:
`)
		flags.PrintDefaults()
		fmt.Fprintf(inout.Stderr, `
Examples:
  $ pumlseqexpand ./seq.puml
  @startuml
  participant A
  participant B
  A -> B: req
  A -> B: inner
  @enduml

  $ pumlseqexpand ./seq.puml | pumlseqrefinement -map ./map.tsv ./abstract.puml -

  $ find . -name '*.puml' -print0 | xargs -0 -n1 pumlseqexpand >/dev/null
`)
	}

	var base string
	flags.StringVar(&base, "base", "", "directory ref's relative paths are resolved from when reading stdin (default: current directory)")
	var name string
	flags.StringVar(&name, "name", "", "diagram name to give the output's @startuml (default: keeps the input's name)")
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

	remaining := flags.Args()
	switch {
	case len(remaining) > 1:
		return nil, fmt.Errorf("too many arguments: accepts at most 1 file path")
	case len(remaining) == 1 && remaining[0] != "-":

		if base != "" {
			return nil, errors.New("-base is only for stdin: it conflicts with a file argument")
		}
		f, err := os.Open(remaining[0])
		if err != nil {
			return nil, err
		}
		return &Options{CommonOptions: &commonOptions, Base: filepath.Dir(remaining[0]), Name: name, Input: tools.NamedInput{Name: remaining[0], Reader: f}}, nil
	default:
		return &Options{CommonOptions: &commonOptions, Base: base, Name: name, Input: tools.NamedInput{Name: "-", Reader: inout.Stdin}}, nil
	}
}
