package main

import (
	"context"
	"github.com/locngoxuan/buildpack"
	"github.com/locngoxuan/buildpack/common"
	"os"
	"os/signal"
	"path/filepath"
)

var version = "1.1.0"

func main() {
	buildpack.Version = version
	arg, err := buildpack.ReadArguments()
	if err != nil {
		common.PrintLog("read argument fail: %v", err)
		os.Exit(1)
	}

	if buildpack.CommandWithoutConfig(arg.Command) {
		bp := buildpack.BuildPack{
			Arguments: arg,
		}
		_ = bp.Run(nil)
		return
	}

	workDir, err := filepath.Abs(".")
	if err != nil {
		common.PrintLog("lookup working directory fail: %v", err)
		os.Exit(1)
	}

	cf := arg.ConfigFile
	if common.IsEmptyString(cf) {
		cf = filepath.Join(workDir, buildpack.ConfigFileName)
	}

	config, err := buildpack.ReadConfig(cf)
	if err != nil {
		common.PrintLog("read buildpack config fail: %v", err)
		os.Exit(1)
	}

	bp, err := buildpack.CreateBuildPack(arg, config)
	if err != nil {
		common.PrintLog("init buildpack fail: %v", err)
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
			bp.Exist(ctx)
			break
		}
	}(signalChannel)

	err = bp.Run(ctx)
	if err != nil {
		common.PrintLog("ERROR: %v", err)
		os.Exit(1)
	}
}
