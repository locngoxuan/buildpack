package builder

import (
	"fmt"
	"github.com/jhoonb/archivex"
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
	moduleInCommon := filepath.Join(ctx.GetCommonDirectory(), ctx.Name)
	err := os.MkdirAll(moduleInCommon, 0777)
	if err != nil {
		return err
	}

	appYMLSrc := filepath.Join(ctx.WorkingDir, wesAppConfig)
	appYMLDst := filepath.Join(moduleInCommon, wesAppConfig)
	buildpack.LogVerbose(ctx.BuildPack, fmt.Sprintf("Copying %s to %s", appYMLSrc, appYMLDst))
	err = buildpack.CopyFile(appYMLSrc, appYMLDst)
	if err != nil {
		return err
	}

	dockerSrc := filepath.Join(ctx.WorkingDir, appDockerfile)
	dockerDst := filepath.Join(moduleInCommon, appDockerfile)
	buildpack.LogVerbose(ctx.BuildPack, fmt.Sprintf("Copying %s to %s", appDockerfile, dockerDst))
	err = buildpack.CopyFile(dockerSrc, dockerDst)
	if err != nil {
		return err
	}

	distFolderSrc := filepath.Join(ctx.WorkingDir, distFolderName)
	distFolderDst := filepath.Join(moduleInCommon, distFolderName)
	err = os.MkdirAll(distFolderDst, 0755)
	if err != nil {
		return err
	}
	err = copyDirectory(ctx.BuildPack, distFolderSrc, distFolderDst)
	if err != nil {
		return err
	}

	libFolderSrc := filepath.Join(ctx.WorkingDir, libsFolderName)
	libFolderDst := filepath.Join(moduleInCommon, libsFolderName)
	err = os.MkdirAll(libFolderDst, 0755)
	if err != nil {
		return err
	}
	err = copyDirectory(ctx.BuildPack, libFolderSrc, libFolderDst)
	if err != nil {
		return err
	}

	pomSrc := filepath.Join(ctx.WorkingDir, "target", pomFileName)
	pomDst := filepath.Join(moduleInCommon, pomFileName)
	buildpack.LogVerbose(ctx.BuildPack, fmt.Sprintf("Copying %s to %s", pomSrc, pomDst))
	err = buildpack.CopyFile(pomSrc, pomDst)
	if err != nil {
		return err
	}

	// tar info
	pom, err := buildpack.ReadPOM(pomSrc)
	if err != nil {
		return err
	}
	tarName := fmt.Sprintf("%s-%s.tar", pom.ArtifactId, ctx.Version)
	tarFile := filepath.Join(ctx.WorkingDir, "target", tarName)
	//create tar
	tar := new(archivex.TarFile)
	err = tar.Create(tarFile)
	if err != nil {
		return err
	}
	err = tar.AddAll(moduleInCommon, false)
	if err != nil {
		return err
	}
	err = tar.Close()
	if err != nil {
		return err
	}

	tarDst := filepath.Join(moduleInCommon, tarName)
	buildpack.LogVerbose(ctx.BuildPack, fmt.Sprintf("Copying %s to %s", tarName, tarDst))
	err = buildpack.CopyFile(tarFile, tarDst)
	if err != nil {
		return err
	}
	return nil
}
