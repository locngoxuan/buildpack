package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
)

var version = "2.0.0"
var workDir string
var arg Arguments
var cfg ProjectConfig

func main() {
	var err error
	arg, err = readArguments()
	if err != nil {
		log.Printf("FAILURE: reading arguments get error %v\n", err)
		os.Exit(1)
	}

	err = readEnvVariables(arg.ConfigFile)
	if err != nil {
		log.Printf("FAILURE: reading arguments get error %v\n", err)
		os.Exit(1)
	}

	workDir, err = filepath.Abs(".")
	if err != nil {
		log.Printf("FAILURE: looking working directory get error %v\n", err)
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
		log.Printf("FAILURE: %s", err)
		os.Exit(1)
	}
}
