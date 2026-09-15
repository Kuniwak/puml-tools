package main

import (
	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/tools/pumlseqfold/cmd"
)

func main() {
	cli.Run(cmd.MainCommandByArgs)
}
