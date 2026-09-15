package main

import (
	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/tools/pumlseqmsg/cmd"
)

func main() {
	cli.Run(cmd.MainCommandByArgs)
}
