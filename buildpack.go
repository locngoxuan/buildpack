package buildpack

import (
	"context"
)

type BuildPack struct {
}

func CreateBuildpack() BuildPack {
	return BuildPack{}
}

func (bp *BuildPack) Run(ctx context.Context) {

}
