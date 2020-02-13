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
		DockerHosts: ctx.BuildPack.Config.DockerConfig.Hosts,
		Version:     ctx.Version,
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
	config, err := sqlbundle.ReadBundle(b.Bundle.BundleFile)
	if err != nil {
		return err
	}
	finalName := fmt.Sprintf("%s-%s-%s.tar", config.Build.Group, config.Build.Artifact, ctx.Version)
	finalBuild := filepath.Join(ctx.WorkingDir, "target", finalName)

	moduleInCommon := filepath.Join(ctx.GetCommonDirectory(), ctx.Name)
	err = os.MkdirAll(moduleInCommon, 0777)
	if err != nil {
		return err
	}

	//copy tar
	published := filepath.Join(moduleInCommon, finalName)
	err = buildpack.CopyFile(finalBuild, published)
	if err != nil {
		return err
	}
	buildpack.LogInfo(ctx.BuildPack, fmt.Sprintf("Copy %s to %s", finalBuild, published))

	//copy bundle
	bundleSrc := filepath.Join(ctx.WorkingDir, "target", bundleFileName)
	bundleDst := filepath.Join(moduleInCommon, bundleFileName)
	err = buildpack.CopyFile(bundleSrc, bundleDst)
	if err != nil {
		return err
	}
	buildpack.LogInfo(ctx.BuildPack, fmt.Sprintf("Copy %s to %s", bundleSrc, bundleDst))
	return nil
}
