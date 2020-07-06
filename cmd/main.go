package main

import (
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack"
	"scm.wcs.fortna.com/lngo/buildpack/common"
)

func main() {
	arg, err := buildpack.ReadArguments()
	if err != nil {
		common.PrintFatal(err, "can not read arguments")
	}

	if buildpack.CommandWithoutConfig(arg.Command) {
		bp, err := buildpack.CreateBuildPack(arg, buildpack.BuildConfig{})
		if err != nil {
			common.PrintFatal(err, "can not init buildpack")
		}
		_ = bp.Run(nil)
		return
	}

	workDir, err := filepath.Abs(".")
	if err != nil {
		common.PrintFatal(err, "can not get current path of working directory")
	}

	cf := arg.ConfigFile
	if common.IsEmptyString(cf) {
		cf = filepath.Join(workDir, buildpack.ConfigFileName)
	}

	common.PrintInfo("get build configuration from %s", cf)
	config, err := buildpack.ReadConfig(cf)
	if err != nil {
		common.PrintFatal(err, "can not read config")
	}

	common.PrintInfo("%v %v %v", arg, config)
	bp, err := buildpack.CreateBuildPack(arg, config)
	if err != nil {
		common.PrintFatal(err, "can not init buildpack")
	}
	err = bp.Run(nil)
	if err != nil {
		common.PrintFatal(err, "fail!")
	}
}
