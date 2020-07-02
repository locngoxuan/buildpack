package buildpack

import (
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack/builder"
)

type Module struct {
	Id   int
	Name string
	Path string

	workDir string
	tmpDir  string
}

type SortedById []Module

func (a SortedById) Len() int           { return len(a) }
func (a SortedById) Less(i, j int) bool { return a[i].Id < a[j].Id }
func (a SortedById) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func (m *Module) start(bp BuildPack) error {
	/**
	1. Read configuration
		- Read Buildpackfile.build
		- Read Buildoackfile.publish
	2. Clean
		- Clean result of build
		- Clean result of publish
		- Clean .buildpack/{module-name}
	3. Build
		- Pre build
		- Build
		- Post build
	4. Publish
		- Pre publish
		- Publish
		- Post publish
	5. Clean (Allow skip)
		- Clean result of build
		- Clean result of publish
		- Clean .buildpack/{module-name}
	 */
	m.workDir = filepath.Join(bp.WorkDir, m.Path)
	m.tmpDir = filepath.Join(bp.WorkDir, BuildPackTmpDir, m.Name)
	bc, err := builder.ReadConfig(m.workDir)
	if err != nil {
		return err
	}
	b, err := builder.GetBuilder(bc.Builder)
	if err != nil {
		return err
	}
	buildContext := builder.BuildContext{
		WorkDir: m.workDir,
		TmpDir:  m.tmpDir,
	}
	err = b.PreBuild(buildContext)
	if err != nil {
		return err
	}
	err = b.Build(buildContext)
	if err != nil {
		return err
	}
	err = b.PostBuild(buildContext)
	if err != nil {
		return err
	}
	return nil
}
