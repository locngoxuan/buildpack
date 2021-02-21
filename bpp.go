package main

import (
	"context"
	"fmt"
)

func run(ctx context.Context) error{
	switch arg.Command {
	case cmdVersion:
		return showVersion()
	case cmdClean:
		return clean(ctx)
	case cmdBuild:
		return build(ctx)
	case cmdHelp:
		f.Usage()
		return nil
	}
	return fmt.Errorf("can recognize command %s", arg.Command)
}
