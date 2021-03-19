package builder

import (
	"fmt"
	"github.com/locngoxuan/sqlbundle"
	"path/filepath"
)

type Sql struct {
}

func (b Sql) Clean(ctx BuildContext) error {
	sqlbundle.SetLogWriter(ctx.LogWriter)
	bundle, err := sqlbundle.NewSQLBundle(sqlbundle.Argument{
		Version: ctx.Version,
		WorkDir: ctx.WorkDir,
	})
	if err != nil {
		return err
	}
	return bundle.Clean()
}

func (b Sql) PreBuild(ctx BuildContext) error {
	return nil
}

func (b Sql) PostFail(ctx BuildContext) error {
	return b.Clean(ctx)
}

func (b Sql) Build(ctx BuildContext) error {
	sqlbundle.SetLogWriter(ctx.LogWriter)
	bundle, err := sqlbundle.NewSQLBundle(sqlbundle.Argument{
		Version: ctx.Version,
		WorkDir: ctx.WorkDir,
	})
	if err != nil {
		return err
	}
	return bundle.Pack()
}

func (b Sql) PostBuild(ctx BuildContext) error {
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

	// copy src directory to common dir
	srcFolder := filepath.Join(ctx.WorkDir, "src")
	if common.Exists(srcFolder) {
		destination := filepath.Join(ctx.OutputDir, "src")
		err = common.CreateDir(common.CreateDirOption{
			AbsPath:       destination,
			SkipContainer: true,
			Perm:          0755,
		})
		if err != nil {
			return err
		}
		err = common.CopyDirectory(ctx.LogWriter, srcFolder, destination)
		if err != nil {
			return err
		}
	}

	//copy deps
	depsFolder := filepath.Join(ctx.WorkDir, "deps")
	if common.Exists(depsFolder) {
		destination := filepath.Join(ctx.OutputDir, "deps")
		err = common.CreateDir(common.CreateDirOption{
			AbsPath:       destination,
			SkipContainer: true,
			Perm:          0755,
		})
		if err != nil {
			return err
		}
		err = common.CopyDirectory(ctx.LogWriter, depsFolder, destination)
		if err != nil {
			return err
		}
	}

	c, err := ReadMvnConfig(ctx.WorkDir)
	if err != nil {
		return err
	}
	return copyUsingFilter(ctx.WorkDir, ctx.OutputDir, c.Config.Filters)
}
