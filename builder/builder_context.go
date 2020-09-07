package builder

import (
	"context"
	"io"
)

type BuildContext struct {
	Name         string
	Path         string
	WorkDir      string
	OutputDir    string
	ShareDataDir string
	Version      string

	SkipContainer  bool
	ContainerImage string
	SkipClean      bool
	LogWriter      io.Writer
	Ctx            context.Context
}
