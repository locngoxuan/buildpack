package main

import (
	"context"
	"fmt"
	"github.com/locngoxuan/buildpack/v1/common"
	"log"
	"os"
	"os/signal"
	"path/filepath"
)

var version = "2.0.0"
var workDir string
var arg Arguments
var cfg BuildConfig

func main() {
	var err error
	arg, err = ReadArguments()
	if err != nil {
		log.Println()
		common.PrintLog("read argument fail: %v", err)
		os.Exit(1)
	}

	err = ReadEnv(arg.ConfigFile)
	if err != nil {
		common.PrintLog("read argument fail: %v", err)
		os.Exit(1)
	}

	workDir, err = filepath.Abs(".")
	if err != nil {
		common.PrintLog("lookup working directory fail: %v", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, GetSingal()...)
	defer func() {
		signal.Stop(signalChannel)
		cancel()
	}()
	go func(ch chan os.Signal) {
		for {
			_ = <-ch
			signal.Stop(ch)
			cancel()
			break
		}
	}(signalChannel)

	err = run(ctx)
	if err != nil {
		fmt.Printf("FAILURE: %s", err)
		os.Exit(1)
	}
}
