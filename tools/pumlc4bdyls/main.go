package main

import (
	"github.com/Kuniwak/puml-tools/cli"
	"github.com/Kuniwak/puml-tools/tools/pumlc4bdyls/cmd"
)

func main() {
	cli.Run(cmd.MainCommandByArgs)
}
