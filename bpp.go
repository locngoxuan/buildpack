package buildpack

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
	case cmdHelp:
		f.Usage()
		return nil
	case cmdClean:
		return clean(ctx)
	case cmdBuild:
		err := prepareConfig()
		if err != nil {
			return err
		}
		return build(ctx)
	case cmdPack:
		err := prepareConfig()
		if err != nil {
			return err
		}
		return pack(ctx)
	case cmdPublish:
		err := prepareConfig()
		if err != nil {
			return err
		}
		return publish(ctx)
	case cmdPump:
		err := prepareConfig()
		if err != nil {
			return err
		}
		return pump(ctx)
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
	//due to each module may have different label. Then it takes version without label here and let appending label
	//is executed in each module
	buildVersion = arg.Version
	if utils.IsStringEmpty(buildVersion) {
		buildVersion = utils.Trim(cfg.Version)
	}
	if utils.IsStringEmpty(buildVersion) {
		return fmt.Errorf("can not recognize version of project")
	}
	return nil
}
