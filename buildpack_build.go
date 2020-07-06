package buildpack

import (
	"errors"
	"fmt"
	"github.com/gosuri/uiprogress"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"os"
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack/common"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

const BuildPackOutputDir = ".buildpack"

var steps = []string{
	"read config ",
	"clean       ",
	"pre build   ",
	"building    ",
	"post build  ",
	"read config ",
	"pre publish ",
	"publishing  ",
	"post publish",
	"completed   ",
}

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
	//normalize index table
	reverseIndexTable := make(map[int]int)
	currentLevel := ms[0].Id
	currentIndex := 0
	for _, m := range ms {
		if currentLevel < m.Id {
			//change level
			currentLevel = m.Id
			currentIndex += 1
		}
		reverseIndexTable[m.Id] = currentIndex
	}

	tree := make(map[int]*sync.WaitGroup)
	var wg sync.WaitGroup

	for _, m := range ms {
		curIndex := reverseIndexTable[m.Id]
		w, ok := tree[curIndex]
		if !ok {
			w = new(sync.WaitGroup)
			tree[curIndex] = w
		}
		w.Add(1)
		wg.Add(1)
	}

	uiprogress.Start()
	summaries := make(map[string]*ModuleSummary)
	var errorCount int32 = 0
	for _, m := range ms {
		summaries[m.Name] = &ModuleSummary{
			Name:   m.Name,
			Result: "STARTED",
		}
		go func(module Module) {
			progress := make(chan int)
			name := module.Name
			if len(name) <= 10 {
				name = fmt.Sprintf("%-10v", name)
			} else {
				name = fmt.Sprintf("%-10v", name[0:10])
			}
			bar := uiprogress.AddBar(len(steps))
			bar.AppendCompleted()
			bar.PrependFunc(func(b *uiprogress.Bar) string {
				if b.Current() == 0 {
					return fmt.Sprintf("[%s] waiting     ", name)
				}
				return fmt.Sprintf("[%s] "+steps[b.Current()-1], name)
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

			//progress <- 1
			moduleIndex := reverseIndexTable[module.Id]
			//get prev wait group
			w, ok := tree[moduleIndex-1]
			if ok {
				w.Wait()
			}

			//continue to build if not found any error
			if atomic.LoadInt32(&errorCount) == 0 {
				e := module.start(*bp, progress)
				if e != nil {
					atomic.AddInt32(&errorCount, 1)
					summaries[module.Name].Result = "ERROR"
					summaries[module.Name].Message = e.Error()
					progress <- -1
				} else {
					summaries[module.Name].Result = "OK"
				}
			} else {
				summaries[module.Name].Result = "ABORTED"
				progress <- -1
			}

			//get current wait group
			w, ok = tree[moduleIndex]
			if ok {
				w.Done()
			}
		}(m)
	}
	wg.Wait()
	uiprogress.Stop()

	common.SetLogOutput(os.Stdout)
	common.PrintInfo("")
	common.PrintInfo("")

	t := table.NewWriter()
	t.SetStyle(table.StyleColoredGreenWhiteOnBlack)
	t.SetTitle("Summary")
	t.Style().Title.Align = text.AlignCenter
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Module", "Result", "Message"})
	for _, e := range summaries {
		t.AppendRow(table.Row{
			e.Name,
			e.Result,
			e.Message,
		})
	}
	t.Render()

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
