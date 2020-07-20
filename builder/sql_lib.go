package builder

import (
	"fmt"
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack/common"
	"github.com/locngoxuan/sqlbundle"
)

type SqlLib struct {
	Sql
}

func (b SqlLib) PostBuild(ctx BuildContext) error {
	packageJsonFile := filepath.Join(ctx.WorkDir, sqlbundle.PACKAGE_JSON)
	packageJson, err := sqlbundle.ReadPackageJSON(packageJsonFile)
	if err != nil {
		return err
	}

	buildDir := filepath.Join(ctx.WorkDir, "build")

	//copy package.json
	jsonSrc := filepath.Join(buildDir, "package", sqlbundle.PACKAGE_JSON)
	jsonPublished := filepath.Join(ctx.OutputDir, sqlbundle.PACKAGE_JSON)
	err = common.CopyFile(jsonSrc, jsonPublished)
	if err != nil {
		return err
	}

	//copy tar
	tarName := fmt.Sprintf("%s.tar", packageJson.ArtifactId)
	tarSrc := filepath.Join(buildDir, tarName)
	tarPublished := filepath.Join(ctx.OutputDir, tarName)
	err = common.CopyFile(tarSrc, tarPublished)
	if err != nil {
		return err
	}
	return nil
}
