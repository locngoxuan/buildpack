package main

import (
	"context"
	"fmt"
	"github.com/locngoxuan/buildpack/config"
	"github.com/locngoxuan/buildpack/utils"
	"log"
	"os"
	"os/signal"
	"path/filepath"
)

var version = "2.0.0"
var workDir string
var outputDir string
var arg Arguments
var cfg config.ProjectConfig
var buildVersion string

func main() {
	var err error
	arg, err = readArguments()
	if err != nil {
		log.Printf("FAILURE: reading arguments get error %v", err)
		os.Exit(1)
	}

	err = readEnvVariables()
	if err != nil {
		log.Printf("FAILURE: reading arguments get error %v", err)
		os.Exit(1)
	}

	workDir, err = filepath.Abs(".")
	if err != nil {
		log.Printf("FAILURE: looking working directory get error %v", err)
		os.Exit(1)
	}
	if !utils.IsStringEmpty(arg.ConfigFile) {
		workDir, _ = filepath.Split(arg.ConfigFile)
	}
	outputDir = filepath.Join(workDir, config.OutputDir)
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
		fmt.Println(fmt.Sprintf("%s: %s", utils.TextRed("FAILURE"), err))
		os.Exit(1)
	}
	fmt.Println(utils.TextGreen("SUCCESS"))
}
