package buildpack

import (
	"context"
	"errors"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"os"
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack/common"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	BuildPackOutputDir = ".buildpack"
	progressCompl      = 0
	progressIncr       = 1
	progressError      = 2
	progressAbort      = 3

	progressStarted     = progressIncr
	progressClean       = progressIncr
	progressPreBuild    = progressIncr
	progressBuild       = progressIncr
	progressPostBuild   = progressIncr
	progressPrePublish  = progressIncr
	progressPublish     = progressIncr
	progressPostPublish = progressIncr
)

var (
	stepWithoutPublish = []string{
		"waiting     ",
		"started     ",
		"clean       ",
		"pre build   ",
		"on build    ",
		"post build  ",
		"completed   ",
	}

	stepWithPublish = []string{
		"waiting      ",
		"started      ",
		"clean        ",
		"pre build    ",
		"on build     ",
		"post build   ",
		"pre publish  ",
		"publishing   ",
		"post publish ",
		"completed    ",
	}
)

type RunModuleOption struct {
	Module
	*Tracker
	*Progress
	reverseIndex map[int]int
	treeWait     map[int]*sync.WaitGroup
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

func renderSummaryTable(summaries []*Tracker, started time.Time) {
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
			e.Result.Result,
			fmt.Sprintf("%.2f s", e.TimeElapsed.Seconds()),
			e.Message,
		})
	}
	t.AppendFooter(table.Row{
		"",
		"Time Elapsed",
		fmt.Sprintf("%.2f s", time.Since(started).Seconds()),
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

	//start progress observation
	pool, f := createPoolOfProgressBar(bp.SkipProgressBar)
	progress := &Progress{
		Wait:         new(sync.WaitGroup),
		errorCount:   0,
		started:      time.Now(),
		Pool:         pool,
		ProgressFunc: f,
		Steps:        stepWithPublish,
		Trackers:     make([]*Tracker, 0),
	}
	if bp.IsSkipPublish() {
		progress.Steps = stepWithoutPublish
	}
	progress.start()
	for _, m := range ms {
		tracker := progress.addTracker(m.Name, getLogFile(*bp, m.Name))

		go runModule(ctx, *bp, RunModuleOption{
			treeWait:     tree,
			reverseIndex: reverseIndexTable,
			Module:       m,
			Tracker:      tracker,
			Progress:     progress,
		})
	}
	//wait to complete
	err = progress.stop()
	//+phase: end build
	common.PrintLog("") //break line

	//render table of result
	renderSummaryTable(progress.Trackers, progress.started)
	//end table render

	if ctx.Err() != nil {
		return errors.New("terminated")
	}

	//return error if error count larger than zero
	if err != nil {
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

func runModule(ctx context.Context, bp BuildPack, option RunModuleOption) {
	module := option.Module
	reverseIndexTable := option.reverseIndex
	tree := option.treeWait
	option.Progress.observer(option.Tracker)
	moduleIndex := reverseIndexTable[module.Id]
	//get prev wait group
	w, ok := tree[moduleIndex-1]
	if ok {
		w.Wait()
	}
	//continue to build if not found any error
	if !option.Progress.isError() || ctx.Err() == nil {
		e := module.start(ctx, bp, option.Tracker)
		if e != nil {
			option.Progress.setErr()
			option.Tracker.onError(e)
		} else if ctx.Err() != nil {
			option.Tracker.onAborted()
		} else {
			option.Tracker.onDone()
		}
	} else {
		option.Tracker.onAborted()
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
