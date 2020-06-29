package buildpack

import "flag"

var f *flag.FlagSet

type Arguments struct {
	Version string
}

type Command struct {
	Cmd string
	Arguments
}

func ParseCommand() Command {

	return Command{}
}
