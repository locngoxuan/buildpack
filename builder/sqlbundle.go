package builder

import (
	"os"
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack/sqlbundle"
)

const (
	sqlBundleBuildTool = "sqlbundle"
	bundleFileName     = "sqlbundle.yml"
)

type SQLBundleBuildTool struct {
	Bundle sqlbundle.SQLBundle
}

func (b *SQLBundleBuildTool) Name() string {
	return sqlBundleBuildTool
}

func (b *SQLBundleBuildTool) GenerateConfig(ctx BuildContext) error {
	return nil
}

func (b *SQLBundleBuildTool) LoadConfig(ctx BuildContext) error {
	b.Bundle = sqlbundle.SQLBundle{
		WorkingDir:  ctx.WorkingDir,
		BundleFile:  filepath.Join(ctx.WorkingDir, bundleFileName),
		Clean:       true,
		Dockerize:   true,
		DockerHosts: ctx.BuildPack.Config.DockerConfig.Hosts,
		Verbose:     ctx.Verbose(),
	}
	return nil
}

func (b *SQLBundleBuildTool) Clean(ctx BuildContext) error {
	return b.Bundle.RunClean()
}

func (b *SQLBundleBuildTool) PreBuild(ctx BuildContext) error {
	return nil
}

func (b *SQLBundleBuildTool) Build(ctx BuildContext) error {
	if ctx.Verbose() {
		return b.Bundle.Run(os.Stdout)
	} else {
		return b.Bundle.Run(nil)
	}
}

func (b *SQLBundleBuildTool) PostBuild(ctx BuildContext) error {
	return nil
}
