package publisher

import (
	"context"
	"io"
)

type PublishContext struct {
	Name      string
	Path      string
	WorkDir   string
	OutputDir string
	RepoName  string
	Version   string
	IsStable  bool
	LogWriter io.Writer
	Ctx       context.Context
}
