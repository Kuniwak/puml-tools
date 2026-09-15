package main

import (
	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/tools/pumlseq2c4/cmd"
)

func main() {
	cli.Run(cmd.MainCommandByArgs)
}
