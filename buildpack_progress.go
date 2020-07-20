package buildpack

import (
	"fmt"
	"github.com/cheggaaa/pb"
	"github.com/pkg/errors"
	"scm.wcs.fortna.com/lngo/buildpack/common"
	"sync"
	"sync/atomic"
	"time"
)

type Progress struct {
	Pool       *pb.Pool
	Wait       *sync.WaitGroup
	Steps      []string
	errorCount int32
	started    time.Time
	ProgressFunc
	Trackers []*Tracker
}

type Tracker struct {
	Name    string
	Signal  chan int
	Bar     *pb.ProgressBar
	started time.Time
	Result
}

type Result struct {
	Result      string
	Message     string
	TimeElapsed time.Duration
	LogFile     string
}

type ProgressFunc func(t *Tracker, steps []string)

func displayProgressBar(t *Tracker, steps []string) {
	for {
		i := <-t.Signal
		if i == progressIncr {
			t.Bar.Increment()
			t.Bar.BarStart = steps[t.Bar.Get()]
		} else if i == progressError {
			t.Bar.BarStart = "error       "
			break
		} else if i == progressAbort {
			t.Bar.BarStart = "aborted     "
			break
		} else {
			t.Bar.Set(len(steps))
			t.Bar.BarStart = steps[len(steps)-1]
			break
		}
	}
	t.Bar.Finish()
}

func displayProgressText(t *Tracker, steps []string) {
	currentStep := 0
	for {
		i := <-t.Signal
		if i == progressIncr {
			currentStep++
			common.PrintLog("module [%s] change step to [%s]", t.Name, steps[currentStep])
		} else if i == progressError {
			common.PrintLog("module [%s] is [error]", t.Name)
			break
		} else if i == progressAbort {
			common.PrintLog("module [%s] is [aborted]", t.Name)
			break
		} else {
			common.PrintLog("module [%s] is [completed]", t.Name)
			break
		}
	}
}

func (b *Progress) createProgressBar(name string, count int) *pb.ProgressBar {
	bar := pb.New(count).Prefix(name)
	bar.SetRefreshRate(100 * time.Millisecond)
	bar.ShowPercent = true
	bar.ShowBar = true
	bar.ShowCounters = false
	bar.ShowElapsedTime = false
	bar.ShowFinalTime = false
	bar.ShowTimeLeft = false
	bar.BarStart = b.Steps[0]
	bar.ShowSpeed = false
	return bar
}

func createPoolOfProgressBar(skipUi bool) (*pb.Pool, ProgressFunc) {
	if skipUi {
		return nil, displayProgressText
	}
	return pb.NewPool(), displayProgressBar
}

func (b *Progress) start() {
	if b.Pool != nil {
		_ = b.Pool.Start()
	}
}

func (b *Progress) observer(tracker *Tracker) {
	go func(t *Tracker, steps []string) {
		b.ProgressFunc(t, steps)
		b.Wait.Done()
	}(tracker, b.Steps)
}

func (b *Progress) setErr() {
	atomic.AddInt32(&b.errorCount, 1)
}

func (b *Progress) isError() bool {
	return atomic.LoadInt32(&b.errorCount) != 0
}

func (b *Progress) addTracker(name, logFile string) *Tracker {
	b.Wait.Add(1)
	t := &Tracker{
		Name:    name,
		Signal:  make(chan int, 1),
		started: time.Now(),
		Result: Result{
			Result:      "STARTED",
			TimeElapsed: 0,
			LogFile:     logFile,
		},
	}
	if b.Pool != nil {
		shortedName := name
		if len(name) <= 10 {
			shortedName = fmt.Sprintf("%-12v", name)
		} else {
			shortedName = fmt.Sprintf("%-12v", name[0:10])
		}
		t.Bar = b.createProgressBar(shortedName, len(b.Steps))
		b.Pool.Add(t.Bar)
	}
	b.Trackers = append(b.Trackers, t)
	return t
}

func (b *Progress) stop() error {
	b.Wait.Wait()
	if b.Pool != nil {
		_ = b.Pool.Stop()
	}
	if b.isError() {
		return errors.New("")
	}
	return nil
}

func (t *Tracker) onStart() {
	t.started = time.Now()
}

func (t *Tracker) onError(e error) {
	t.Result.Result = "ERROR"
	t.Result.Message = fmt.Sprintf("%v. Detail at %s", e, t.Result.LogFile)
	t.Result.TimeElapsed = time.Since(t.started)
	t.Signal <- progressError
}

func (t *Tracker) onDone() {
	t.Result.Result = "DONE"
	t.Result.TimeElapsed = time.Since(t.started)
	t.Signal <- progressCompl
}

func (t *Tracker) onAborted() {
	t.Result.Result = "ABORTED"
	t.Signal <- progressAbort
}
