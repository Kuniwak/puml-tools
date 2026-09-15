package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/pumlseqast"
	"github.com/Kuniwak/puml-tools/pumlseqtsv"
	"github.com/Kuniwak/puml-tools/tools"
)

type Options struct {
	CommonOptions *tools.CommonOptions
	MsgKeys       map[pumlseqast.MessageKey]bool
	Name          string
	Reader        io.ReadCloser
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("pumlseqhide", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		fmt.Fprintf(inout.Stderr, `Usage: pumlseqhide -msg-file <path|-> [options] (<file>|< <file>)

Removes, line by line, every message in a sequence diagram that matches the message set given
by -msg-file (this hides the internal interaction as tau, i.e. CSP hiding P \ A). The message
set is a TSV of (sender<TAB>receiver<TAB>label); the output of pumlseqmsg can be piped in
directly.
Matching is on the (sender, receiver, label) triple; arrow style is ignored. Any note attached
to a hidden message (one immediately following it, attached to that participant) is removed
along with it. The note's own text is not part of the match.
ref over is never hidden and is left untouched (to remove an internal ref, first run
pumlseqexpand to turn it into real messages).
Control blocks left empty by the removal (empty alt/par branches, or empty
opt/loop/group/break, etc.) are pruned automatically, and a par/alt reduced to a single branch
has its contents hoisted to the parent level (recursively, from the innermost block out, and
idempotently).
Pass - to -msg-file to read the message set from stdin (in that case, pass the diagram as a
file argument instead).
Pass -name to set the name on the output's @startuml (omit it to keep the input's name).
For PNG input, the embedded PlantUML metadata is decoded.

Options:
`)
		flags.PrintDefaults()
		fmt.Fprintf(inout.Stderr, `
Examples:
  $ pumlseqhide --msg-file <(pumlseqmsg --common B,C ./seq.puml) ./seq.puml
  $ pumlseqmsg --common B,C ./seq.puml | pumlseqhide --msg-file - ./seq.puml
  $ pumlseqhide --msg-file - -name Hidden ./seq.puml <./msgs.tsv
`)
	}

	var msgFile string
	flags.StringVar(&msgFile, "msg-file", "", "TSV of messages to hide (sender<TAB>receiver<TAB>label); - reads from stdin")
	var name string
	flags.StringVar(&name, "name", "", "name to set on the output's @startuml (omit to keep the input's name)")
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

	if msgFile == "" {
		return nil, errors.New("-msg-file is required")
	}

	remaining := flags.Args()
	if len(remaining) > 1 {
		return nil, fmt.Errorf("too many arguments: accepts at most 1 file path")
	}
	diagramFromStdin := !(len(remaining) == 1 && remaining[0] != "-")

	keys, err := readMsgKeys(msgFile, inout, diagramFromStdin)
	if err != nil {
		return nil, err
	}

	if !diagramFromStdin {
		f, err := os.Open(remaining[0])
		if err != nil {
			return nil, err
		}
		return &Options{CommonOptions: &commonOptions, MsgKeys: keys, Name: name, Reader: f}, nil
	}
	return &Options{CommonOptions: &commonOptions, MsgKeys: keys, Name: name, Reader: inout.Stdin}, nil
}

func readMsgKeys(msgFile string, inout *cli.ProcInout, diagramFromStdin bool) (map[pumlseqast.MessageKey]bool, error) {
	if msgFile == "-" {
		if diagramFromStdin {
			return nil, errors.New("cannot read both the diagram and -msg-file from stdin")
		}
		return pumlseqtsv.ParseMessageKeys(inout.Stdin)
	}
	f, err := os.Open(msgFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return pumlseqtsv.ParseMessageKeys(f)
}
