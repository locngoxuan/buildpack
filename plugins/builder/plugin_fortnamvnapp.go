package main

import (
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack/builder"
	"scm.wcs.fortna.com/lngo/buildpack/common"
)

type MvnApp struct {
	builder.Mvn
}

const (
	wesAppConfig   = "application.yml"
	distFolderName = "dist"
	libsFolderName = "libs"
	appDockerfile  = "Dockerfile"
)

func (b MvnApp) PostBuild(ctx builder.BuildContext) error {
	pomSrc := filepath.Join(ctx.WorkDir, "target", builder.PomXML)
	pomDst := filepath.Join(ctx.OutputDir, builder.PomXML)
	err := common.CopyFile(pomSrc, pomDst)
	if err != nil {
		return err
	}

	// copy application.yml to common dir
	appYMLSrc := filepath.Join(ctx.WorkDir, wesAppConfig)
	appYMLDst := filepath.Join(ctx.OutputDir, wesAppConfig)
	err = common.CopyFile(appYMLSrc, appYMLDst)
	if err != nil {
		return err
	}

	// copy Dockerfile to common dir
	dockerSrc := filepath.Join(ctx.WorkDir, appDockerfile)
	if common.Exists(dockerSrc) {
		dockerDst := filepath.Join(ctx.OutputDir, appDockerfile)
		err = common.CopyFile(dockerSrc, dockerDst)
		if err != nil {
			return err
		}
	}

	// copy dist directory to common dir
	distFolderSrc := filepath.Join(ctx.WorkDir, distFolderName)
	distFolderDst := filepath.Join(ctx.OutputDir, distFolderName)
	err = common.CreateDir(common.CreateDirOption{
		SkipContainer: true,
		Perm:          0755,
		AbsPath:       distFolderDst,
	})
	if err != nil {
		return err
	}
	err = common.CopyDirectory(ctx.LogWriter, distFolderSrc, distFolderDst)
	if err != nil {
		return err
	}

	// copy libs directory to common dir

	libFolderSrc := filepath.Join(ctx.WorkDir, libsFolderName)
	if common.Exists(libFolderSrc) {
		libFolderDst := filepath.Join(ctx.OutputDir, libsFolderName)
		err = common.CreateDir(common.CreateDirOption{
			SkipContainer: true,
			Perm:          0755,
			AbsPath:       libFolderDst,
		})
		if err != nil {
			return err
		}
		err = common.CopyDirectory(ctx.LogWriter, libFolderSrc, libFolderDst)
		if err != nil {
			return err
		}
	}
	return nil
}