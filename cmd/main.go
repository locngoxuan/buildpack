package main

import (
	"flag"
	"os"
	. "scm.wcs.fortna.com/lngo/buildpack"
)

const version = "0.1.0"

func main() {
	f := flag.NewFlagSet("buildpack", flag.ContinueOnError)
	f.Usage = func() {
		/**
		Do nothing
		 */
	}

	if len(os.Args) <= 1 {
		Usage(f)
		return
	}
	action := os.Args[1]
	runtimeConfig, err := ReadArgument(f)
	if err != nil {
		Usage(f)
		return
	}

	if runtimeConfig.IsHelp() {
		Usage(f)
		return
	}

	configFile := FileBuildPackConfig
	if len(runtimeConfig.ConfigFile()) > 0 {
		configFile = runtimeConfig.ConfigFile()
	}
	config, err := ReadFromConfigFile(configFile)
	if err != nil {
		LogFatal(BuildResult{
			Success: false,
			Action:  action,
			Phase:   "init",
			Err:     err,
			Message: "",
		})
		return
	}

	err = verifyAction(action)
	if err != nil {
		LogFatal(BuildResult{
			Success: false,
			Action:  action,
			Phase:   "init",
			Err:     err,
			Message: "",
		})
	}

	buildPack, err := NewBuildPack(action, config, runtimeConfig)
	if err != nil {
		LogFatal(BuildResult{
			Success: false,
			Action:  action,
			Phase:   "init",
			Err:     err,
			Message: "",
		})
	}
	result := Handle(buildPack)
	if !result.Success {
		LogFatal(result)
	}
	os.Exit(0)
}
