package buildpack

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/locngoxuan/buildpack/common"
	"os"
	"path/filepath"
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
  build         Run build and publish to repository
  version       Show version of buildpack
  help          Show usage

Examples:
  buildpack clean
  buildpack version
  buildpack build --dev-mode
  buildpack build --release
  buildpack build --path --skip-progress

Options:
`
)

type Arguments struct {
	Command         string
	Version         string
	Module          string
	ConfigFile      string
	ShareData       string
	LogDir          string
	BuildRelease    bool
	BuildPath       bool
	Verbose         bool
	DevMode         bool
	IncreaseVersion bool
	SkipOption
}

type SkipOption struct {
	SkipContainer   bool
	SkipPublish     bool
	SkipClean       bool
	SkipProgressBar bool

	NoGitTag   bool
	NoBackward bool
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
	f.BoolVar(&arg.SkipPublish, "skip-publish", false, "skip publish build to repository")
	f.BoolVar(&arg.SkipProgressBar, "skip-progress", false, "use text plain instead of progress ui")

	//git operation
	f.BoolVar(&arg.IncreaseVersion, "increase-version", false, "force to increase version after build")
	f.BoolVar(&arg.NoGitTag, "no-git-tag", false, "skip tagging source code")
	f.BoolVar(&arg.NoBackward, "no-backward", false, "if true, then major version will be increased")

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

func ReadEnv(configFile string) error {
	workDir, err := filepath.Abs(".")
	if err != nil {
		return err
	}

	if !common.IsEmptyString(configFile) {
		workDir, _ = filepath.Split(configFile)
	}

	envFile := filepath.Join(workDir, ".env")
	if !common.Exists(envFile) {
		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		envFile = filepath.Join(userHomeDir, ".buildpack", ".env")
	}

	if !common.Exists(envFile) {
		return nil
	}

	f, err := os.Open(envFile)
	if err != nil {
		return err
	}

	defer func() {
		_ = f.Close()
	}()

	fi, err := f.Stat()
	if err != nil {
		return err
	}

	if fi.IsDir() {
		return nil
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			continue
		}
		err = os.Setenv(parts[0], parts[1])
		if err != nil {
			return err
		}
	}
	return nil
}
