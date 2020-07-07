package buildpack

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack/builder"
	"scm.wcs.fortna.com/lngo/buildpack/common"
	"scm.wcs.fortna.com/lngo/buildpack/publisher"
	"time"
)

type Module struct {
	Id   int
	Name string
	Path string
}

type ModuleSummary struct {
	Name        string
	Result      string
	Message     string
	TimeElapsed time.Duration
	LogFile     string
}

type SortedById []Module

func (a SortedById) Len() int           { return len(a) }
func (a SortedById) Less(i, j int) bool { return a[i].Id < a[j].Id }
func (a SortedById) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func (m Module) clean(bp BuildPack) error {
	workDir := filepath.Join(bp.WorkDir, m.Path)
	outputDir := filepath.Join(bp.WorkDir, BuildPackOutputDir, m.Name)
	//create log writer
	logFile := filepath.Join(bp.WorkDir, BuildPackOutputDir, fmt.Sprintf("%s.log", m.Name))
	if !common.IsEmptyString(bp.LogDir) {
		logFile = filepath.Join(bp.LogDir, fmt.Sprintf("%s.log", m.Name))
	}
	file, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()
	//clean
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
		LogWriter:     file,
	}
	err = b.Clean(buildContext)
	if err != nil {
		return err
	}

	_ = common.DeleteDir(common.DeleteDirOption{
		SkipContainer: true,
		AbsPath:       logFile,
	})
	return nil
}

func (m Module) start(ctx context.Context, bp BuildPack, progress chan<- int) error {
	if ctx.Err() != nil {
		return nil
	}
	progress <- progressStarted
	workDir := filepath.Join(bp.WorkDir, m.Path)
	outputDir := filepath.Join(bp.WorkDir, BuildPackOutputDir, m.Name)

	//create log writer
	logFile := filepath.Join(bp.WorkDir, BuildPackOutputDir, fmt.Sprintf("%s.log", m.Name))
	if !common.IsEmptyString(bp.LogDir) {
		logFile = filepath.Join(bp.LogDir, fmt.Sprintf("%s.log", m.Name))
	}
	file, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()
	//create version of build
	v := bp.GetVersion()
	//begin build phase
	if ctx.Err() != nil {
		return nil
	}

	bc, err := builder.ReadConfig(workDir)
	if err != nil {
		_, _ = fmt.Fprintln(file, fmt.Sprintf("read build config get error %v", err))
		return err
	}
	if !bp.BuildRelease && !bp.BuildPath {
		// it means build with label
		v = fmt.Sprintf("%s-%s", bp.GetVersion(), bc.Label)
	}
	b, err := builder.GetBuilder(bc.Builder)
	if err != nil {
		_, _ = fmt.Fprintln(file, fmt.Sprintf("can not fild builder. error: %v", err))
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
		LogWriter:     file,
	}

	if ctx.Err() != nil {
		return nil
	}
	progress <- progressClean
	err = b.Clean(buildContext)
	if err != nil {
		return err
	}

	if ctx.Err() != nil {
		return nil
	}
	progress <- progressPreBuild
	err = b.PreBuild(buildContext)
	if err != nil {
		_, _ = fmt.Fprintf(file, "pre build get error %v\n", err)
		return err
	}

	if ctx.Err() != nil {
		return nil
	}
	progress <- progressBuild
	err = b.Build(buildContext)
	if err != nil {
		_, _ = fmt.Fprintf(file, "build get error %v\n", err)
		_ = b.PostFail(buildContext)
		return err
	}

	if ctx.Err() != nil {
		return nil
	}
	progress <- progressPostBuild
	err = b.PostBuild(buildContext)
	if err != nil {
		_, _ = fmt.Fprintf(file, "post build get error %v\n", err)
		return err
	}

	if ctx.Err() != nil {
		return nil
	}
	if !bp.IsSkipClean() {
		err = b.Clean(buildContext)
		if err != nil {
			_, _ = fmt.Fprintf(file, "clean after build get error %v\n", err)
			return err
		}
	}
	//end build phase

	//begin publish phase
	if ctx.Err() != nil {
		return nil
	}
	if bp.IsSkipPublish() {
		_ = common.DeleteDir(common.DeleteDirOption{
			SkipContainer: true,
			AbsPath:       logFile,
		})
		return nil
	}
	//end publish phase

	progress <- progressPrePublish
	if ctx.Err() != nil {
		return nil
	}
	pc, err := publisher.ReadConfig(workDir)
	if err != nil {
		_, _ = fmt.Fprintf(file, "reaed publish config get error %v\n", err)
		return err
	}
	p, err := publisher.GetPublisher(pc.Publisher)
	if err != nil {
		_, _ = fmt.Fprintf(file, "find publisher get error %v\n", err)
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
		LogWriter: file,
	}

	if ctx.Err() != nil {
		return nil
	}
	err = p.PrePublish(publishCtx)
	if err != nil {
		_, _ = fmt.Fprintf(file, "pre publish get error %v\n", err)
		return err
	}

	if ctx.Err() != nil {
		return nil
	}
	progress <- progressPublish
	err = p.Publish(publishCtx)
	if err != nil {
		_, _ = fmt.Fprintf(file, "publish get error %v\n", err)
		return err
	}

	if ctx.Err() != nil {
		return nil
	}
	progress <- progressPostPublish
	err = p.PostPublish(publishCtx)
	if err != nil {
		_, _ = fmt.Fprintf(file, "post publish get error %v\n", err)
		return err
	}

	_ = common.DeleteDir(common.DeleteDirOption{
		SkipContainer: true,
		AbsPath:       logFile,
	})
	return nil
}
