package cmd

import (
	"sort"

	"github.com/Kuniwak/puml-tools/cli"
	pumlc42seqchkcmd "github.com/Kuniwak/puml-tools/tools/pumlc42seqchk/cmd"
	pumlc4bdylscmd "github.com/Kuniwak/puml-tools/tools/pumlc4bdyls/cmd"
	pumlc4foldcmd "github.com/Kuniwak/puml-tools/tools/pumlc4fold/cmd"
	pumlparsecmd "github.com/Kuniwak/puml-tools/tools/pumlparse/cmd"
	pumlseq2c4cmd "github.com/Kuniwak/puml-tools/tools/pumlseq2c4/cmd"
	pumlseqbreakcmd "github.com/Kuniwak/puml-tools/tools/pumlseqbreak/cmd"
	pumlseqexpandcmd "github.com/Kuniwak/puml-tools/tools/pumlseqexpand/cmd"
	pumlseqfoldcmd "github.com/Kuniwak/puml-tools/tools/pumlseqfold/cmd"
	pumlseqgroupcutcmd "github.com/Kuniwak/puml-tools/tools/pumlseqgroupcut/cmd"
	pumlseqgroupmaskcmd "github.com/Kuniwak/puml-tools/tools/pumlseqgroupmask/cmd"
	pumlseqgroupsplitcmd "github.com/Kuniwak/puml-tools/tools/pumlseqgroupsplit/cmd"
	pumlseqhidecmd "github.com/Kuniwak/puml-tools/tools/pumlseqhide/cmd"
	pumlseqmsgcmd "github.com/Kuniwak/puml-tools/tools/pumlseqmsg/cmd"
	pumlseqprojcmd "github.com/Kuniwak/puml-tools/tools/pumlseqproj/cmd"
	pumlseqrefinementcmd "github.com/Kuniwak/puml-tools/tools/pumlseqrefinement/cmd"
	pumlseqsubsetcmd "github.com/Kuniwak/puml-tools/tools/pumlseqsubset/cmd"
	pumlseqtracecmd "github.com/Kuniwak/puml-tools/tools/pumlseqtrace/cmd"
	pumltodocmd "github.com/Kuniwak/puml-tools/tools/pumltodo/cmd"
)

type ToolEntry struct {
	Name    string
	Summary string
	Hidden  bool
	Help    func(args []string, inout *cli.ProcInout) int
}

var Registry = []ToolEntry{
	{Name: "pumlparse", Summary: "Extracts the PlantUML embedded in PNG metadata, parses it to an AST, and renders it back to text (round-trip).", Help: pumlparsecmd.MainCommandByArgs},
	{Name: "pumlseqgroupcut", Summary: "Cuts out the group block with the given label from a sequence diagram.", Help: pumlseqgroupcutcmd.MainCommandByArgs},
	{Name: "pumlseqgroupmask", Summary: "Removes the group block with the given label from a sequence diagram.", Help: pumlseqgroupmaskcmd.MainCommandByArgs},
	{Name: "pumlseqgroupsplit", Summary: "Splits a sequence diagram whose messages are all wrapped in groups into one *.puml file per group.", Help: pumlseqgroupsplitcmd.MainCommandByArgs},
	{Name: "pumltodo", Summary: "Extracts TODOs from PlantUML comments and notes and lists them as a 3-column TSV.", Help: pumltodocmd.MainCommandByArgs},
	{Name: "pumlseqbreak", Summary: "Detects and lists messages in a sequence diagram whose causal chain is broken.", Help: pumlseqbreakcmd.MainCommandByArgs},
	{Name: "pumlseqfold", Summary: "Folds several participants of a sequence diagram into one (summarization).", Help: pumlseqfoldcmd.MainCommandByArgs},
	{Name: "pumlseqmsg", Summary: "Extracts the common (synchronized) events of a given participant set and prints them as a TSV.", Help: pumlseqmsgcmd.MainCommandByArgs},
	{Name: "pumlseqhide", Summary: "Removes, line by line, every message matching the given message set (hiding).", Help: pumlseqhidecmd.MainCommandByArgs},
	{Name: "pumlseqproj", Summary: "Keeps only the messages involving the given participants (projection).", Help: pumlseqprojcmd.MainCommandByArgs},
	{Name: "pumlseqsubset", Summary: "Checks whether two sequence diagrams have the same message trace (or one is a subsequence of the other).", Help: pumlseqsubsetcmd.MainCommandByArgs},
	{Name: "pumlseqrefinement", Summary: "Checks whether a sequence diagram is a refinement of another (CSP trace refinement).", Help: pumlseqrefinementcmd.MainCommandByArgs},
	{Name: "pumlseqexpand", Summary: "Replaces each ref over in a sequence diagram with the body of the diagram it references (expansion).", Help: pumlseqexpandcmd.MainCommandByArgs},
	{Name: "pumlseqtrace", Summary: "Enumerates a sequence diagram's event traces and writes each one as a single-column TSV file.", Help: pumlseqtracecmd.MainCommandByArgs},
	{Name: "pumlseq2c4", Summary: "Builds a single C4-PlantUML diagram from the union of several sequence diagrams.", Help: pumlseq2c4cmd.MainCommandByArgs},
	{Name: "pumlc42seqchk", Summary: "Checks a sequence diagram against a C4 diagram for consistency.", Help: pumlc42seqchkcmd.MainCommandByArgs},
	{Name: "pumlc4fold", Summary: "Folds several elements of a C4 diagram into one (for example, a C4 level 2 Container diagram back down to level 1).", Help: pumlc4foldcmd.MainCommandByArgs},
	{Name: "pumlc4bdyls", Summary: "Lists the elements contained in a C4 diagram's boundary as a TSV with a header.", Help: pumlc4bdylscmd.MainCommandByArgs},

	{Name: "pumlhelp", Summary: "Prints the help of every PlantUML tool, gathered together as Markdown.", Hidden: true},
}

func init() {
	for i := range Registry {
		if Registry[i].Name == "pumlhelp" {
			Registry[i].Help = MainCommandByArgs
		}
	}
}

func KnownTool(registry []ToolEntry, name string) bool {
	for _, entry := range registry {
		if entry.Name == name {
			return true
		}
	}
	return false
}

func SelectEntries(registry []ToolEntry, names []string) []ToolEntry {
	requested := make(map[string]bool, len(names))
	for _, name := range names {
		requested[name] = true
	}
	keep := func(entry ToolEntry) bool {
		if len(names) == 0 {
			return !entry.Hidden
		}
		return requested[entry.Name]
	}

	var entries []ToolEntry
	for _, entry := range registry {
		if keep(entry) {
			entries = append(entries, entry)
		}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	return entries
}
