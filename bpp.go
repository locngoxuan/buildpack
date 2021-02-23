package main

import (
	"context"
	"fmt"
	"github.com/locngoxuan/buildpack/config"
	"github.com/locngoxuan/buildpack/utils"
)

func run(ctx context.Context) error {
	switch arg.Command {
	case cmdVersion:
		return showVersion()
	case cmdClean:
		return clean(ctx)
	case cmdBuild:
		//read configuration
		err := prepareConfig()
		if err != nil {
			return err
		}
		return build(ctx)
	case cmdHelp:
		f.Usage()
		return nil
	}
	return fmt.Errorf("can recognize command %s", arg.Command)
}

func prepareConfig() error {
	var err error
	cfg, err = config.ReadProjectConfig(workDir, arg.ConfigFile)
	if err != nil {
		return nil
	}
	//initializing version
	buildVersion = arg.Version
	if utils.IsStringEmpty(buildVersion) {
		buildVersion = utils.Trim(cfg.Version)
	}
	if utils.IsStringEmpty(buildVersion) {
		return fmt.Errorf("can not recognize version of project")
	}
	return nil
}
