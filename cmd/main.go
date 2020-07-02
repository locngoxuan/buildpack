package main

import (
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack"
)

func main() {
	arg, err := buildpack.ReadArguments()
	if err != nil {
		buildpack.PrintFatal(err, "can not read arguments")
	}

	if buildpack.CommandWithoutConfig(arg.Command) {
		bp := buildpack.CreateBuildPack(arg, buildpack.Environments{}, buildpack.BuildConfig{})
		_ = bp.Run(nil)
		return
	}
	env, err := buildpack.ReadEnvironment()
	if err != nil {
		buildpack.PrintFatal(err, "can not read environment")
	}

	workDir, err := filepath.Abs(".")
	if err != nil {
		buildpack.PrintFatal(err, "can not get current path of working directory")
	}

	cf := arg.ConfigFile
	if buildpack.IsEmptyString(cf) {
		cf = filepath.Join(workDir, buildpack.ConfigFileName)
	}

	buildpack.PrintInfo("get build configuration from %s", cf)
	config, err := buildpack.ReadConfig(cf)
	if err != nil {
		buildpack.PrintFatal(err, "can not read config")
	}

	buildpack.PrintInfo("%v %v %v", arg, env, config)
	bp := buildpack.CreateBuildPack(arg, env, config)
	err = bp.Run(nil)
	if err != nil {
		buildpack.PrintFatal(err, "build fail")
	}
}
