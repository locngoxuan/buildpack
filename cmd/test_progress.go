package main
//
//import (
//	"github.com/gosuri/uiprogress"
//	"sync"
//	"time"
//)
//
//func main() {
//	//uiprogress.Start() // start rendering
//	//var steps = []string{"downloading source", "installing deps", "compiling", "packaging", "seeding database", "deploying", "staring servers"}
//	//bar := uiprogress.AddBar(len(steps)) // Add a new bar
//	//bar.PrependFunc(func(b *uiprogress.Bar) string {
//	//	return "app: " + steps[b.Current()-1]
//	//})
//	//// optionally, append and prepend completion and elapsed time
//	//bar.AppendCompleted()
//	//bar.PrependElapsed()
//	//
//	//for bar.Incr() {
//	//	time.Sleep(time.Second * 2)
//	//}
//
//	waitTime := time.Millisecond * 100
//	uiprogress.Start()
//
//	// start the progress bars in go routines
//	var wg sync.WaitGroup
//
//	bar1 := uiprogress.AddBar(20).AppendCompleted().PrependElapsed()
//	wg.Add(1)
//	go func() {
//		defer wg.Done()
//		for bar1.Incr() {
//			time.Sleep(waitTime)
//		}
//	}()
//
//	bar2 := uiprogress.AddBar(40).AppendCompleted().PrependElapsed()
//	wg.Add(1)
//	go func() {
//		defer wg.Done()
//		for bar2.Incr() {
//			time.Sleep(waitTime)
//		}
//	}()
//
//	time.Sleep(time.Second)
//	bar3 := uiprogress.AddBar(20).PrependElapsed().AppendCompleted()
//	wg.Add(1)
//	go func() {
//		defer wg.Done()
//		for i := 1; i <= bar3.Total; i++ {
//			bar3.Set(i)
//			time.Sleep(waitTime)
//		}
//	}()
//
//	// wait for all the go routines to finish
//	wg.Wait()
//}
