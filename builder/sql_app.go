package builder

import (
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack/common"
	"scm.wcs.fortna.com/lngo/sqlbundle"
)

type SqlApp struct {
	Sql
}

func (b SqlApp) PostBuild(ctx BuildContext) error {
	buildDir := filepath.Join(ctx.WorkDir, "build")
	//copy package.json
	jsonSrc := filepath.Join(buildDir, "package", sqlbundle.PACKAGE_JSON)
	jsonPublished := filepath.Join(ctx.OutputDir, sqlbundle.PACKAGE_JSON)
	err := common.CopyFile(jsonSrc, jsonPublished)
	if err != nil {
		return err
	}

	// copy src directory to common dir
	srcFolder := filepath.Join(ctx.WorkDir, "src")
	if common.Exists(srcFolder) {
		destination := filepath.Join(ctx.OutputDir, "src")
		err = common.CreateDir(destination, true, 0755)
		if err != nil {
			return err
		}
		err = common.CopyDirectory(srcFolder, destination)
		if err != nil {
			return err
		}
	}

	depsFolder := filepath.Join(ctx.WorkDir, "deps")
	if common.Exists(depsFolder) {
		destination := filepath.Join(ctx.OutputDir, "deps")
		err = common.CreateDir(destination, true, 0755)
		if err != nil {
			return err
		}
		err = common.CopyDirectory(depsFolder, destination)
		if err != nil {
			return err
		}
	}

	// copy Dockerfile to common dir
	dockerSrc := filepath.Join(ctx.WorkDir, appDockerfile)
	if common.Exists(dockerSrc) {
		destination := filepath.Join(ctx.OutputDir, appDockerfile)
		err = common.CopyFile(dockerSrc, destination)
		if err != nil {
			return err
		}
	}
	return nil
}

func init() {
	registries["sql_app"] = &SqlApp{}
}
