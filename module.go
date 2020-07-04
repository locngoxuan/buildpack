package buildpack

import (
	"fmt"
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack/builder"
	"scm.wcs.fortna.com/lngo/buildpack/publisher"
)

type Module struct {
	Id   int
	Name string
	Path string
}

type SortedById []Module

func (a SortedById) Len() int           { return len(a) }
func (a SortedById) Less(i, j int) bool { return a[i].Id < a[j].Id }
func (a SortedById) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func (m *Module) clean(bp BuildPack) error {
	workDir := filepath.Join(bp.WorkDir, m.Path)
	outputDir := filepath.Join(bp.WorkDir, BuildPackOutputDir, m.Name)
	bc, err := builder.ReadConfig(workDir)
	if err != nil {
		return err
	}
	b, err := builder.GetBuilder(bc.Builder)
	if err != nil {
		return err
	}

	v := bp.GetVersion()
	if !bp.BuildRelease && !bp.BuildPath {
		// it means build with label
		v = fmt.Sprintf("%s-%s", bp.GetVersion(), bc.Label)
	}

	buildContext := builder.BuildContext{
		Name:          m.Name,
		Path:          m.Path,
		WorkDir:       workDir,
		OutputDir:     outputDir,
		SkipContainer: bp.IsSkipContainer(),
		SkipClean:     bp.SkipClean,
		ShareDataDir:  bp.ShareData,
		Version:       v,
	}
	err = b.Clean(buildContext)
	if err != nil {
		return err
	}
	return nil
}

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
	workDir := filepath.Join(bp.WorkDir, m.Path)
	outputDir := filepath.Join(bp.WorkDir, BuildPackOutputDir, m.Name)

	//create version of build
	v := bp.GetVersion()

	//begin build phase
	bc, err := builder.ReadConfig(workDir)
	if err != nil {
		return err
	}
	if !bp.BuildRelease && !bp.BuildPath {
		// it means build with label
		v = fmt.Sprintf("%s-%s", bp.GetVersion(), bc.Label)
	}
	b, err := builder.GetBuilder(bc.Builder)
	if err != nil {
		return err
	}
	buildContext := builder.BuildContext{
		Name:          m.Name,
		Path:          m.Path,
		WorkDir:       workDir,
		OutputDir:     outputDir,
		SkipContainer: bp.IsSkipContainer(),
		SkipClean:     bp.SkipClean,
		ShareDataDir:  bp.ShareData,
		Version:       v,
	}
	err = b.Clean(buildContext)
	if err != nil {
		return err
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

	if !bp.IsSkipClean() {
		err = b.Clean(buildContext)
		if err != nil {
			return err
		}
	}
	//end build phase

	//begin publish phase
	if bp.IsSkipPublish() {
		return nil
	}
	//end publish phase
	pc, err := publisher.ReadConfig(workDir)
	if err != nil {
		return err
	}
	p, err := publisher.GetPublisher(pc.Publisher)
	if err != nil {
		return err
	}
	publishCtx := publisher.PublishContext{
		Name:      m.Name,
		Path:      m.Path,
		WorkDir:   workDir,
		OutputDir: outputDir,
		Version:   v,
		RepoName:  pc.Repository,
		IsStable:  bp.BuildRelease || bp.BuildPath,
	}
	err = p.PrePublish(publishCtx)
	if err != nil {
		return err
	}
	err = p.Publish(publishCtx)
	if err != nil {
		return err
	}
	err = p.PostPublish(publishCtx)
	if err != nil {
		return err
	}
	return nil
}
