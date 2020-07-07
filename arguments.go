package buildpack

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

var (
	f       = flag.NewFlagSet("buildpack", flag.ContinueOnError)
	verbose = false

	cmdVersion = "version"
	cmdBuild   = "build"
	cmdClean   = "clean"
	cmdHelp    = "help"

	usagePrefix = `Usage: buildpack COMMAND [OPTIONS]
COMMAND:
  clean         Clean build folder		
  build         Run build phase
  version       Show version of buildpack
  help          Show usage

Examples:
  buildpack clean
  buildpack version
  buildpack build --dev-mode
  buildpack build --release

Options:
`
)

type Arguments struct {
	Command      string
	Version      string
	Module       string
	ConfigFile   string
	ShareData    string
	LogDir       string
	BuildRelease bool
	BuildPath    bool
	Verbose      bool
	DevMode      bool
	SkipOption
}

type SkipOption struct {
	SkipContainer   bool
	SkipPublish     bool
	SkipGit         bool
	SkipBranching   bool
	SkipClean       bool
	SkipProgressBar bool
}

func ReadArguments() (arg Arguments, err error) {
	f.SetOutput(os.Stdout)
	f.StringVar(&arg.Version, "version", "", "version number")
	f.StringVar(&arg.Module, "module", "", "list of module")
	f.StringVar(&arg.ShareData, "share-data", "", "sharing directory")
	f.StringVar(&arg.LogDir, "log-dir", "", "log directory")
	f.StringVar(&arg.ConfigFile, "config", "", "specific path to config file")
	f.BoolVar(&arg.DevMode, "dev-mode", false, "enable local mode to disable container build")
	f.BoolVar(&arg.BuildRelease, "release", false, "build for releasing")
	f.BoolVar(&arg.BuildPath, "patch", false, "build for patching")

	f.BoolVar(&arg.SkipClean, "skip-clean", false, "skip clean everything after build complete")
	f.BoolVar(&arg.SkipContainer, "skip-container", false, "skip container build")
	f.BoolVar(&arg.SkipGit, "skip-git", false, "skip git operation")
	f.BoolVar(&arg.SkipBranching, "skip-branching", false, "skip checkout new branch")
	f.BoolVar(&arg.SkipPublish, "skip-publish", false, "skip publish build to repository")
	f.BoolVar(&arg.SkipProgressBar, "skip-progress", false, "use text plain instead of progress ui")

	f.BoolVar(&verbose, "verbose", false, "show more detail in console")
	f.Usage = func() {
		_, _ = fmt.Fprint(f.Output(), usagePrefix)
		f.PrintDefaults()
		os.Exit(1)
	}
	if len(os.Args) == 1 {
		f.Usage()
		return
	}

	arg.Command = strings.TrimSpace(os.Args[1])
	if len(os.Args) > 2 {
		err = f.Parse(os.Args[2:])
	}
	arg.Verbose = verbose
	return
}
