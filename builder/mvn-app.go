package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack"
)

const (
	mvnAppBuildTool = "mvn-app"
	wesAppConfig    = "application.yml"
	distFolderName  = "dist"
	libsFolderName  = "libs"
	appDockerfile   = "Dockerfile"
)

type MVNAppBuildTool struct {
	MVNBuildTool
}

func (c *MVNAppBuildTool) Name() string {
	return mvnAppBuildTool
}

func (c *MVNAppBuildTool) Build(ctx BuildContext) error {
	return c.MVNBuildTool.Build(ctx)
}

func (c *MVNAppBuildTool) PostBuild(ctx BuildContext) error {
	moduleInCommonDir := filepath.Join(ctx.GetCommonDirectory(), ctx.Name)
	err := os.MkdirAll(moduleInCommonDir, 0777)
	if err != nil {
		return err
	}
	// copy pom.xml to .buildpack/{modulename}/ directory
	pomSrc := filepath.Join(ctx.WorkingDir, "target", pomFileName)
	pomDst := filepath.Join(moduleInCommonDir, pomFileName)
	buildpack.LogVerbose(ctx.BuildPack, fmt.Sprintf("Copying %s to %s", pomSrc, pomDst))
	err = buildpack.CopyFile(pomSrc, pomDst)
	if err != nil {
		return err
	}

	// copy application.yml to common dir
	appYMLSrc := filepath.Join(ctx.WorkingDir, wesAppConfig)
	appYMLDst := filepath.Join(moduleInCommonDir, wesAppConfig)
	buildpack.LogInfo(ctx.BuildPack, fmt.Sprintf("Copying %s to %s", appYMLSrc, appYMLDst))
	err = buildpack.CopyFile(appYMLSrc, appYMLDst)
	if err != nil {
		return err
	}

	// copy Dockerfile to common dir
	dockerSrc := filepath.Join(ctx.WorkingDir, appDockerfile)
	dockerDst := filepath.Join(moduleInCommonDir, appDockerfile)
	buildpack.LogInfo(ctx.BuildPack, fmt.Sprintf("Copying %s to %s", appDockerfile, dockerDst))
	err = buildpack.CopyFile(dockerSrc, dockerDst)
	if err != nil {
		return err
	}

	// copy dist directory to common dir
	distFolderSrc := filepath.Join(ctx.WorkingDir, distFolderName)
	distFolderDst := filepath.Join(moduleInCommonDir, distFolderName)
	buildpack.LogInfo(ctx.BuildPack, fmt.Sprintf("Copying %s to %s", distFolderName, distFolderDst))
	err = os.MkdirAll(distFolderDst, 0755)
	if err != nil {
		return err
	}
	err = copyDirectory(ctx.BuildPack, distFolderSrc, distFolderDst)
	if err != nil {
		return err
	}

	// copy libs directory to common dir
	libFolderSrc := filepath.Join(ctx.WorkingDir, libsFolderName)
	libFolderDst := filepath.Join(moduleInCommonDir, libsFolderName)
	buildpack.LogInfo(ctx.BuildPack, fmt.Sprintf("Copying %s to %s", libsFolderName, libFolderDst))
	err = os.MkdirAll(libFolderDst, 0755)
	if err != nil {
		return err
	}
	err = copyDirectory(ctx.BuildPack, libFolderSrc, libFolderDst)
	if err != nil {
		return err
	}
	return nil
}
