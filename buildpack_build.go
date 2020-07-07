package buildpack

import (
	"context"
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
	"time"
)

const (
	BuildPackOutputDir = ".buildpack"
	progressCompl      = 0
	progressIncr       = 1
	progressError      = 2
	progressAbort      = 3
)

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

func prepareListModule(bp BuildPack) ([]Module, error) {
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
		return nil, errors.New("not found any module")
	}

	//sorting by id
	sort.Sort(SortedById(ms))
	return ms, nil
}

func renderSummaryTable(summaries map[string]*ModuleSummary, started time.Time) {
	t := table.NewWriter()
	t.SetTitle("Summary")
	t.Style().Title.Align = text.AlignCenter
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Module", "Result", "Duration", "Message"})
	t.SetColumnConfigs([]table.ColumnConfig{
		{Name: "Duration", Align: text.AlignRight},
	})
	for _, e := range summaries {
		t.AppendRow(table.Row{
			e.Name,
			e.Result,
			fmt.Sprintf("%d ms", e.TimeElapsed.Milliseconds()),
			e.Message,
		})
	}
	t.AppendFooter(table.Row{
		"",
		"Time Elapsed",
		fmt.Sprintf("%v s", time.Since(started).Seconds()),
		"",
	})
	t.Render()
}

func (bp *BuildPack) build(ctx context.Context) error {
	ms, err := prepareListModule(*bp)
	if err != nil {
		return err
	}
	//create tmp directory
	outputDir := filepath.Join(bp.WorkDir, BuildPackOutputDir)
	err = common.DeleteDir(common.DeleteDirOption{
		SkipContainer: true,
		AbsPath:       outputDir,
	})
	if err != nil {
		return err
	}
	err = common.CreateDir(common.CreateDirOption{
		SkipContainer: true,
		AbsPath:       outputDir,
		Perm:          0755,
	})
	if err != nil {
		return err
	}
	for _, module := range ms {
		err = common.CreateDir(common.CreateDirOption{
			SkipContainer: true,
			AbsPath:       filepath.Join(outputDir, module.Name),
			Perm:          0755,
		})
		if err != nil {
			return err
		}
	}

	//+phase: build
	//normalize index table
	reverseIndexTable := createReverseIndexTable(ms)

	//create tree of wait group
	tree := createTreeWaitGroup(reverseIndexTable, ms)

	//create summary of result
	summaries := make(map[string]*ModuleSummary)

	//create error counter
	var errorCount int32 = 0

	if !bp.SkipProgressBar {
		uiprogress.Start()
	}
	wg := new(sync.WaitGroup)
	started := time.Now()

	var isError int32 = 0
	for _, m := range ms {
		modSum := &ModuleSummary{
			Name:        m.Name,
			Result:      "STARTED",
			TimeElapsed: 0,
			LogFile:     getLogFile(*bp, m.Name),
		}
		summaries[m.Name] = modSum
		wg.Add(1)
		progress := make(chan int)
		go runModule(ctx, *bp, RunModuleOption{
			treeWait:        tree,
			reverseIndex:    reverseIndexTable,
			globalWait:      wg,
			Module:          m,
			ModuleSummary:   modSum,
			skipProgressBar: bp.SkipProgressBar,
			progress:        progress,
		}, &isError)
	}
	wg.Wait()
	if !bp.SkipProgressBar {
		uiprogress.Stop()
	}
	//+phase: end build
	common.PrintLog("") //break line

	//render table of result
	renderSummaryTable(summaries, started)
	//end table render

	if ctx.Err() != nil {
		return errors.New("terminated")
	}

	//return error if error count larger than zero
	if atomic.LoadInt32(&errorCount) > 0 {
		return errors.New("")
	}

	//+phase: git operation
	if bp.IsSkipGit() {
		return nil
	}

	common.PrintLog("") //break line

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

type RunModuleOption struct {
	Module
	*ModuleSummary
	progress        chan int
	reverseIndex    map[int]int
	treeWait        map[int]*sync.WaitGroup
	globalWait      *sync.WaitGroup
	skipProgressBar bool
}

func progressBar(progress chan int, b *uiprogress.Bar, wg *sync.WaitGroup) {
	for {
		i := <-progress
		if i == progressIncr {
			b.Incr()
		} else if i == progressError || i == progressAbort {
			break
		} else {
			for b.Current() < len(steps) {
				b.Incr()
			}
			break
		}
	}
	wg.Done()
}

func progressText(name string, progress chan int, wg *sync.WaitGroup) {
	currentStep := 0
	for {
		i := <-progress
		if i == progressIncr {
			currentStep++
			common.PrintLog("module [%s] change to step [%s]", name, steps[currentStep])
		} else if i == progressError {
			common.PrintLog("module [%s] is [error]", name)
			break
		} else if i == progressAbort {
			common.PrintLog("module [%s] is [aborted]", name)
			break
		} else {
			common.PrintLog("module [%s] is [completed]", name)
			break
		}
	}
	wg.Done()
}

func runModule(ctx context.Context, bp BuildPack, option RunModuleOption, isError *int32) {
	module := option.Module
	skipProgress := option.skipProgressBar
	wg := option.globalWait
	reverseIndexTable := option.reverseIndex
	tree := option.treeWait
	progress := option.progress

	name := module.Name
	if len(name) <= 10 {
		name = fmt.Sprintf("%-10v", name)
	} else {
		name = fmt.Sprintf("%-10v", name[0:10])
	}
	if !skipProgress {
		bar := uiprogress.AddBar(len(steps))
		bar.AppendCompleted()
		bar.PrependFunc(func(b *uiprogress.Bar) string {
			if b.Current() == 0 {
				return fmt.Sprintf("[%s] waiting     ", name)
			}
			return fmt.Sprintf("[%s] "+steps[b.Current()-1], name)
		})
		go progressBar(progress, bar, wg)
	} else {
		go progressText(module.Name, progress, wg)
	}
	//progress <- 1
	moduleIndex := reverseIndexTable[module.Id]
	//get prev wait group
	w, ok := tree[moduleIndex-1]
	if ok {
		w.Wait()
	}

	//continue to build if not found any error
	if atomic.LoadInt32(isError) == 0 || ctx.Err() == nil {
		s := time.Now()
		e := module.start(ctx, bp, progress)
		if e != nil {
			atomic.AddInt32(isError, 1)
			option.ModuleSummary.Result = "ERROR"
			option.ModuleSummary.Message = fmt.Sprintf("%v. Detail at %s", e, option.ModuleSummary.LogFile)
			option.ModuleSummary.TimeElapsed = time.Since(s)
			progress <- progressError
		} else {
			if ctx.Err() != nil {
				option.ModuleSummary.Result = "ABORTED"
				progress <- progressAbort
			} else {
				option.ModuleSummary.Result = "DONE"
				option.ModuleSummary.TimeElapsed = time.Since(s)
				progress <- progressCompl
			}
		}
	} else {
		option.ModuleSummary.Result = "ABORTED"
		progress <- progressAbort
	}

	//get current wait group
	w, ok = tree[moduleIndex]
	if ok {
		w.Done()
	}
}

func getLogFile(bp BuildPack, name string) string {
	file := filepath.Join(bp.WorkDir, BuildPackOutputDir, fmt.Sprintf("%s.log", name))
	if !common.IsEmptyString(bp.LogDir) {
		file = filepath.Join(bp.LogDir, fmt.Sprintf("%s.log", name))
	}
	return file
}

func createReverseIndexTable(ms []Module) map[int]int {
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
	return reverseIndexTable
}

func createTreeWaitGroup(reverseIndexTable map[int]int, ms []Module) map[int]*sync.WaitGroup {
	tree := make(map[int]*sync.WaitGroup)
	for _, m := range ms {
		curIndex := reverseIndexTable[m.Id]
		w, ok := tree[curIndex]
		if !ok {
			w = new(sync.WaitGroup)
			tree[curIndex] = w
		}
		w.Add(1)
	}
	return tree
}

func gitUpdateConfig(cli common.GitClient, bp BuildPack, ver common.Version) (err error) {
	config := bp.BuildConfig
	config.Version = ver.String()
	err = rewriteConfig(config, bp.GetConfigFile())
	if err != nil {
		return fmt.Errorf("rewrite version before do git operation fail %s", err.Error())
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
		return fmt.Errorf("push change of version to git server fail %s", err.Error())
	}
	return nil
}
