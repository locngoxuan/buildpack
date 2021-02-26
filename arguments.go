package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/locngoxuan/buildpack/config"
	"github.com/locngoxuan/buildpack/utils"
	"os"
	"path/filepath"
	"strings"
)

var (
	f = flag.NewFlagSet("BPP", flag.ContinueOnError)

	cmdVersion = "version"
	cmdBuild   = "build"
	cmdPack    = "pack"
	cmdPublish = "publish"
	cmdPump    = "pump"
	cmdClean   = "clean"
	cmdHelp    = "help"

	usagePrefix = `Usage: bpp COMMAND [OPTIONS]
COMMAND:
  clean         Cleaning output of build process

  build         Compiling source code
                (Options: config, release, share-data, module, version, local)

  pack          Packing output of build process as publishable files
                (Options: config, release, module, version, local)

  publish       Publish packages to repository
                (Options: config, module, version)

  pump          Increasing version of project
                (Options: patch, release, skip-backward, force-pull, git-branch)

  version       Showing version of bpp

  help          Showing usage

Examples:
  bpp clean
  bpp version
  bpp build --release --local  
  bpp package --release
  bpp publish
  bpp pump --skip-backward --git-branch=develop    

Options:
`
)

type Arguments struct {
	Command      string
	Version      string
	Module       string
	ConfigFile   string
	ShareData    string
	GitBranch    string
	BuildLocal   bool
	BuildRelease bool
	BuildPath    bool
	ForcePull    bool
	SkipOption
}

type SkipOption struct {
	SkipBackward bool
}

func readArguments() (arg Arguments, err error) {
	f.SetOutput(os.Stdout)
	f.StringVar(&arg.Version, "version", "", "specify version for build")
	f.StringVar(&arg.Module, "module", "", "modules will be built")
	f.StringVar(&arg.ShareData, "share-data", "", "sharing directory for any build and any project on same host")
	f.StringVar(&arg.ConfigFile, "config", "", "specify location of configuration file")
	f.BoolVar(&arg.BuildRelease, "release", false, "project is built for releasing")
	f.BoolVar(&arg.BuildPath, "patch", false, "project is built only for path")
	f.BoolVar(&arg.BuildLocal, "local", false, "running build and clean in local")
	f.StringVar(&arg.GitBranch, "git-branch", "", "branch that code will be pushed")
	f.BoolVar(&arg.SkipBackward, "skip-backward", false, "if true, then major version will be increased")
	f.BoolVar(&arg.ForcePull, "force-pull", false, "if true, then after pumping version, it will try to pulling latest code")

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
	return
}

func updateEnvFromFile(envFile string) error {
	if utils.IsNotExists(envFile) {
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

func readEnvVariables() error {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	err = updateEnvFromFile(filepath.Join(userHomeDir, config.OutputDir, config.ConfigEnvVariables))
	if err != nil {
		return err
	}
	err = updateEnvFromFile(filepath.Join(workDir, config.ConfigEnvVariables))
	if err != nil {
		return err
	}
	return nil
}
