package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack"
)

var (
	usagePrefix = `Usage: buildpack ACTION [OPTIONS]

ACTION:
  init         Init a template of configuration file with name buildpack.yml		
  config       Generate config file for all modules bases on buildpack.yaml
  version      Show the buildpack version information
  clean        Clean working directory
  builder      List all builder supported
  publisher    List all publisher supported
  build        Create a build then publish to repository if --release=true is set

Examples:
  buildpack init -v=0.1.0              
  buildpack config        
  buildpack version
  buildpack build --label=beta         
  buildpack build --release           
  buildpack build --release --path     

Options:`
)

func Usage(f *flag.FlagSet) {
	fmt.Println(usagePrefix)
	f.PrintDefaults()
	os.Exit(1)
}

func main() {
	f := flag.NewFlagSet("buildpack", flag.ContinueOnError)
	f.Usage = func() {
		/**
		Do nothing
		 */
	}

	if len(os.Args) < 2 {
		_, _ = buildpack.ReadForUsage(f)
		Usage(f)
		return
	}

	runtimeConfig, err := buildpack.ReadArgument(f)
	if err != nil {
		Usage(f)
		return
	}

	if runtimeConfig.IsHelp() {
		Usage(f)
		return
	}

	action := os.Args[1]
	err = verifyAction(action)
	if err != nil {
		Usage(f)
		return
	}

	root, err := filepath.Abs(".")
	if err != nil {
		buildpack.LogFatal(buildpack.BuildResult{
			Success: false,
			Action:  action,
			Phase:   "init",
			Err:     err,
			Message: "",
		})
		return
	}

	configFile := filepath.Join(root, buildpack.FileBuildPackConfig)
	if len(runtimeConfig.ConfigFile()) > 0 {
		configFile = runtimeConfig.ConfigFile()
	}
	config, err := buildpack.ReadFromConfigFile(configFile)
	if err != nil && action != actionInit{
		buildpack.LogFatal(buildpack.BuildResult{
			Success: false,
			Action:  action,
			Phase:   "init",
			Err:     err,
			Message: "",
		})
		return
	}

	buildPack, err := buildpack.NewBuildPack(action, config, runtimeConfig)
	if err != nil {
		buildpack.LogFatal(buildpack.BuildResult{
			Success: false,
			Action:  action,
			Phase:   "init",
			Err:     err,
			Message: "",
		})
	}
	result := Handle(buildPack)
	if !result.Success {
		buildpack.LogFatal(result)
	}
	os.Exit(0)
}
