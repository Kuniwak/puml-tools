package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/outfiles"
	"github.com/Kuniwak/puml-tools/pumlseqtrace"
	"github.com/Kuniwak/puml-tools/tools"
)

type Options struct {
	CommonOptions *tools.CommonOptions
	OutDir        string
	Basename      string
	Trace         pumlseqtrace.Options
	Sampling      *float64
	Seed          uint64
	Name          string
	Reader        io.ReadCloser
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("pumlseqtrace", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		fmt.Fprintf(inout.Stderr, `Usage: pumlseqtrace -o <out_dir> [options] (<file>|< <file>)

Enumerates a sequence diagram's event traces and writes each one as a
single-column TSV file. An event is a single message, in the order it
appears in the diagram. Each TSV has one column, the message's label, and
one event per line.
File names follow <out_dir>/<basename>_trace_<idx>.tsv, where idx starts at
1 and is zero-padded to the number of digits in the trace count. Written
paths are printed to stdout, one per line.

Per-block interpretation:
  alt      one trace per compartment.
  opt      two traces: one that runs the body, one that doesn't.
  loop     one trace for each of 1..<loop-unroll> repetitions
           (-loop-unroll 0 unrolls to 0 repetitions only).
  par      enumerates the full interleaving of its operands, keeping the
           order within each operand.
  break    runs up to the body and then cuts the trace short (nothing
           after it runs). A break inside a par cuts the whole trace short
           at the point that event occurs, so events still pending in
           other, concurrent operands never run either.
  group and other blocks with no control structure are transparent.

par is prone to a trace-count explosion. When the count is expected to
exceed %d, a warning is printed to stderr; interrupt with Ctrl+C and add
-interleave-sampling if you want to cut it down (enumeration continues
after the warning either way).
Pass -interleave-sampling <rate> to sample traces down, keeping each one
with probability rate.
The seed is fixed, so the same input and arguments always produce the same
result (override it with -seed).

basename defaults to the input file name with a trailing .png or .puml
stripped. When reading from stdin, -basename is required, and a -basename
containing a path separator (/ \) is an error.
<out_dir> is created if it does not exist, and files with the same name
are overwritten. Every file is written to a temporary file first and only
then renamed into place, so a failure while generating or writing content
leaves <out_dir> untouched.
An unexpanded ref over is an error (a call to another diagram can't be
represented as a single event, so run pumlseqexpand first).
For PNG input, the embedded PlantUML metadata is read.

Options:
`, warnLimit)
		flags.PrintDefaults()
		fmt.Fprintf(inout.Stderr, `
Examples:
  $ pumlseqtrace -o ./out ./seq.puml
  ./out/seq_trace_1.tsv
  ./out/seq_trace_2.tsv

  $ pumlseqtrace -o ./out -basename flow -interleave-sampling 0.01 <./seq.puml
  ./out/flow_trace_1.tsv
`)
	}

	var outDir string
	flags.StringVar(&outDir, "o", "", "directory to write the trace TSV files into (created if it does not exist)")
	var basename string
	flags.StringVar(&basename, "basename", "", "base name for output files (default: derived from the input file name; required when reading from stdin)")
	var loopUnroll int
	flags.IntVar(&loopUnroll, "loop-unroll", 1, "maximum number of loop unrollings (enumerates each of 1..N repetitions; 0 skips the body entirely)")
	var sampling float64
	flags.Float64Var(&sampling, "interleave-sampling", 0, "probability of keeping each trace (0 < rate <= 1; unset means no sampling)")
	var seed uint64
	flags.Uint64Var(&seed, "seed", 1, "seed for the pseudo-random generator used for sampling")
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

	if outDir == "" {
		return nil, errors.New("-o is required")
	}
	if loopUnroll < 0 {
		return nil, fmt.Errorf("-loop-unroll must not be negative: %d", loopUnroll)
	}

	var samplingOption *float64
	flags.Visit(func(f *flag.Flag) {
		if f.Name == "interleave-sampling" {
			samplingOption = &sampling
		}
	})
	if samplingOption != nil && (*samplingOption <= 0 || *samplingOption > 1) {
		return nil, fmt.Errorf("-interleave-sampling must be in (0, 1]: %v", *samplingOption)
	}

	remaining := flags.Args()
	if len(remaining) > 1 {
		return nil, fmt.Errorf("too many arguments: accepts at most 1 file path")
	}

	filePath := ""
	if len(remaining) == 1 && remaining[0] != "-" {
		filePath = remaining[0]
	}
	resolved, err := outfiles.ResolveBasename(basename, filePath)
	if err != nil {
		if errors.Is(err, outfiles.ErrBasenameEmpty) {
			return nil, errors.New("-basename is required when reading the diagram from stdin")
		}
		return nil, fmt.Errorf("-basename %q: %w", basename, err)
	}

	name := inputName(filePath)
	reader, err := cli.OpenInput(name, inout)
	if err != nil {
		return nil, err
	}
	return &Options{
		CommonOptions: &commonOptions,
		OutDir:        outDir,
		Basename:      resolved,
		Trace:         pumlseqtrace.Options{LoopUnroll: loopUnroll},
		Sampling:      samplingOption,
		Seed:          seed,
		Name:          name,
		Reader:        reader,
	}, nil
}

func inputName(filePath string) string {
	if filePath == "" {
		return "-"
	}
	return filePath
}
