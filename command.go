package buildpack

type Command struct {
	Cmd string
	Arguments
	Environments
}

func ParseCommand() Command {
	return Command{}
}
