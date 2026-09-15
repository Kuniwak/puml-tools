package main

import (
	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/tools/pumlseqbreak/cmd"
)

func main() {
	cli.Run(cmd.MainCommandByArgs)
}
