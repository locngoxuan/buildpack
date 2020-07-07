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
	"time"
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
	err := common.DeleteDir(common.DeleteDirOption{
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
		//err := common.CreateDir(filepath.Join(outputDir, module.Name), true, 0755)
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
	var errorCount int32 = 0

	if !bp.SkipProgressBar {
		uiprogress.Start()
	}
	wg := new(sync.WaitGroup)
	started := time.Now()
	for _, m := range ms {
		modSum := &ModuleSummary{
			Name:        m.Name,
			Result:      "STARTED",
			TimeElapsed: 0,
			LogFile:     getLogFile(*bp, m.Name),
		}
		summaries[m.Name] = modSum
		wg.Add(1)
		go runModule(*bp, RunModuleOption{
			treeWait:        tree,
			reverseIndex:    reverseIndexTable,
			globalWait:      wg,
			Module:          m,
			ModuleSummary:   modSum,
			errorCount:      errorCount,
			skipProgressBar: bp.SkipProgressBar,
		})
	}
	wg.Wait()
	if !bp.SkipProgressBar {
		uiprogress.Stop()
	}
	//+phase: end build
	common.PrintLog("") //break line

	//render table of result
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
	//end table render
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
	reverseIndex    map[int]int
	treeWait        map[int]*sync.WaitGroup
	globalWait      *sync.WaitGroup
	skipProgressBar bool
	errorCount      int32
	*ModuleSummary
}

func progressBar(progress chan int, b *uiprogress.Bar, wg *sync.WaitGroup) {
	for {
		i := <-progress
		if i == 1 {
			b.Incr()
		} else if i == -1 || i == -2 {
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
		if i == 1 {
			currentStep++
			common.PrintLog("module [%s] change to step [%s]", name, steps[currentStep])
		} else if i == - 1 {
			common.PrintLog("module [%s] is [error]", name)
			break
		} else if i == - 2 {
			common.PrintLog("module [%s] is [aborted]", name)
			break
		} else {
			common.PrintLog("module [%s] is [completed]", name)
			break
		}
	}
	wg.Done()
}

func runModule(bp BuildPack, option RunModuleOption) {
	module := option.Module
	skipProgress := option.skipProgressBar
	wg := option.globalWait
	reverseIndexTable := option.reverseIndex
	tree := option.treeWait

	progress := make(chan int)
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
	if atomic.LoadInt32(&option.errorCount) == 0 {
		s := time.Now()
		e := module.start(bp, progress)
		if e != nil {
			atomic.AddInt32(&option.errorCount, 1)
			//summaries[module.Name].Result = "ERROR"
			//summaries[module.Name].Message = fmt.Sprintf("%v. Detail at %s", e, summaries[module.Name].LogFile)
			//summaries[module.Name].TimeElapsed = time.Since(s)

			option.ModuleSummary.Result = "ERROR"
			option.ModuleSummary.Message = fmt.Sprintf("%v. Detail at %s", e, option.ModuleSummary.LogFile)
			option.ModuleSummary.TimeElapsed = time.Since(s)
			progress <- -1
		} else {
			//summaries[module.Name].Result = "DONE"
			//summaries[module.Name].TimeElapsed = time.Since(s)
			option.ModuleSummary.Result = "DONE"
			option.ModuleSummary.TimeElapsed = time.Since(s)
			progress <- 0
		}
	} else {
		//summaries[module.Name].Result = "ABORTED"
		option.ModuleSummary.Result = "ABORTED"
		progress <- -2
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
