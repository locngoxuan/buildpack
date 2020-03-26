package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack"
	"scm.wcs.fortna.com/lngo/buildpack/sqlbundle"
)

const (
	sqlBundleBuildTool = "sqlbundle"
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
	bundleFile, err := ctx.GetFile(sqlbundle.FileConfig())
	if err != nil {
		return err
	}
	b.Bundle = sqlbundle.SQLBundle{
		WorkingDir: ctx.WorkingDir,
		BundleFile: bundleFile,
		Clean:      true,
		Version:    ctx.Version,
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
	moduleInCommon := filepath.Join(ctx.GetCommonDirectory(), ctx.Name)
	err := os.MkdirAll(moduleInCommon, 0777)
	if err != nil {
		return err
	}
	sqlTarget, err := ctx.GetFile("target")
	if err != nil {
		return err
	}
	err = copyDirectory(ctx.BuildPack, sqlTarget, moduleInCommon)
	if err != nil {
		return err
	}

	dockerSrc, err := ctx.GetFile(appDockerfile)
	if err != nil {
		return err
	}
	dockerDst := filepath.Join(moduleInCommon, appDockerfile)
	buildpack.LogInfo(ctx.BuildPack, fmt.Sprintf("Copying %s to %s", appDockerfile, dockerDst))
	err = buildpack.CopyFile(dockerSrc, dockerDst)
	if err != nil {
		return err
	}
	return nil
}
