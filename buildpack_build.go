package buildpack

import (
	"errors"
	"fmt"
	"github.com/gosuri/uiprogress"
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack/common"
	"sort"
	"strings"
	"sync"
)

const BuildPackOutputDir = ".buildpack"

func (bp *BuildPack) build() error {
	ms := make([]Module, 0)
	if common.IsEmptyString(bp.Arguments.Module) {
		for _, module := range bp.BuildConfig.Modules {
			ms = append(ms, Module{
				Id:   module.Id,
				Name: module.Name,
				Path: module.Path,
			})
		}
	} else {
		modules := strings.Split(bp.Arguments.Module, ",")
		mmap := make(map[string]struct{})
		for _, module := range modules {
			mmap[module] = struct{}{}
		}

		for _, module := range bp.BuildConfig.Modules {
			if _, ok := mmap[module.Name]; !ok {
				continue
			}
			ms = append(ms, Module{
				Id:   module.Id,
				Name: module.Name,
				Path: module.Path,
			})
		}
	}

	if len(ms) == 0 {
		return errors.New("not found any module")
	}

	//sorting by id
	sort.Sort(SortedById(ms))

	//create tmp directory
	outputDir := filepath.Join(bp.WorkDir, BuildPackOutputDir)
	err := common.DeleteDir(outputDir, true)
	if err != nil {
		return err
	}

	err = common.CreateDir(outputDir, true, 0755)
	if err != nil {
		return err
	}
	for _, module := range ms {
		err := common.CreateDir(filepath.Join(outputDir, module.Name), true, 0755)
		if err != nil {
			return err
		}
	}

	//build
	var wg sync.WaitGroup
	uiprogress.Start()
	var steps = []string{"read build config", "clean", "pre build", "building", "post build", "read publish config", "pre publish", "publishing", "post publish", "completed"}
	for _, m := range ms {
		progress := make(chan int)
		wg.Add(1)
		go func(module Module) {
			bar := uiprogress.AddBar(len(steps))
			bar.AppendCompleted().PrependElapsed()
			bar.PrependFunc(func(b *uiprogress.Bar) string {
				if b.Current() == 0 {
					return fmt.Sprintf("%s: ", module.Name)
				}
				return fmt.Sprintf("%s: "+steps[b.Current()-1], module.Name)
			})
			go func(b *uiprogress.Bar) {
				for {
					i := <-progress
					if i == 1 {
						bar.Incr()
					} else if i == -1 {
						break
					} else {
						for bar.Current() < len(steps) {
							bar.Incr()
						}
						break
					}
				}
				wg.Done()
			}(bar)
			err = module.start(*bp, progress)
			if err != nil {
				progress <- -1
			}
		}(m)
	}
	wg.Wait()
	uiprogress.Stop()
	//git operation
	if bp.IsSkipGit() {
		return nil
	}

	cli := common.GetGitClient()
	defer cli.Close()

	ver, err := common.Parse(bp.GetVersion())
	if err != nil {
		return err
	}
	ver.NextPatch()
	if !bp.IsSkipGitBraching() {
		err = gitUpdateConfig(cli, *bp, ver)
		if err != nil {
			return err
		}
		//create new branch
		branch := ver.MinorBranch()
		err = cli.CreateNewBranch(branch)
		if err != nil {
			return err
		}
	}

	if bp.BuildRelease {
		ver.NextMinor()
	}
	err = gitUpdateConfig(cli, *bp, ver)
	if err != nil {
		return err
	}
	return nil
}

func gitUpdateConfig(cli common.GitClient, bp BuildPack, ver common.Version) (err error) {
	config := bp.BuildConfig
	config.Version = ver.String()
	err = rewriteConfig(config, bp.ConfigFile)
	if err != nil {
		return err
	}

	// push to repo
	err = cli.Add(ConfigFileName)
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("[BUILDPACK] Pump version from %s to %s", bp.GetVersion(), ver.String())
	err = cli.Commit(msg)
	if err != nil {
		return err
	}
	err = cli.Push()
	if err != nil {
		return err
	}
	return nil
}
