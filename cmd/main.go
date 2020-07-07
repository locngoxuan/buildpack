package main

import (
	"os"
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack"
	"scm.wcs.fortna.com/lngo/buildpack/common"
)

func main() {
	arg, err := buildpack.ReadArguments()
	if err != nil {
		common.PrintLog("read argument fail: %v", err)
		os.Exit(1)
	}

	if buildpack.CommandWithoutConfig(arg.Command) {
		bp, err := buildpack.CreateBuildPack(arg, buildpack.BuildConfig{})
		if err != nil {
			common.PrintLog("init buildpack fail: %v", err)
			os.Exit(1)
		}
		_ = bp.Run(nil)
		return
	}

	workDir, err := filepath.Abs(".")
	if err != nil {
		common.PrintLog("lookup working directry fail: %v", err)
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
	err = bp.Run(nil)
	if err != nil {
		common.PrintLog("ERROR: %v", err)
		os.Exit(1)
	}
}
