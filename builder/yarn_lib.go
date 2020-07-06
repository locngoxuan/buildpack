package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack/common"
)

type YarnLib struct {
	Yarn
}

func (b YarnLib) PostBuild(ctx BuildContext) error {
	err := b.yarnPack(ctx)
	if err != nil {
		return err
	}
	//copy package.json -> ./buildpack/test/package.json
	jsonFile := filepath.Join(ctx.WorkDir, packageJson)
	err = common.CopyFile(jsonFile, filepath.Join(ctx.OutputDir, packageJson))
	if err != nil {
		return err
	}

	//rollback version of package.json
	jsonFileBackup := filepath.Join(ctx.WorkDir, packageJsonBackup)
	_ = os.RemoveAll(jsonFile)
	err = common.CopyFile(jsonFileBackup, jsonFile)
	if err != nil {
		return err
	}
	err = common.DeleteDir(common.DeleteDirOption{
		SkipContainer: ctx.SkipContainer,
		AbsPath: filepath.Join(ctx.WorkDir, jsonFileBackup),
		WorkDir: ctx.WorkDir,
		RelativePath: jsonFileBackup,
	})
	if err != nil {
		return err
	}
	config, err := common.ReadNodeJSPackageJson(filepath.Join(ctx.WorkDir, packageJson))
	if err != nil {
		return err
	}
	//copy {name}-{version}.tgz -> ./buildpack/test/{name}-v{version}.tgz
	tgzName := fmt.Sprintf("%s.tgz", config.Name)
	tgzSource := filepath.Join(ctx.WorkDir, tgzName)
	return common.CopyFile(tgzSource, filepath.Join(ctx.OutputDir, tgzName))
}

func init() {
	registries["yarn_lib"] = &YarnLib{}
}
