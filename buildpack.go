package buildpack

import (
	"context"
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack/common"
)

const version = "2.0.0"

type BuildPack struct {
	WorkDir string

	Arguments
	Environments
	BuildConfig

	RepoManager
	GitManager
}

func (bp BuildPack) IsSkipContainer() bool {
	if bp.DevMode {
		return true
	}
	return bp.Arguments.SkipContainer
}

func CommandWithoutConfig(cmd string) bool {
	switch cmd {
	case cmdVersion,
		cmdHelp:
		return true
	case cmdBuild,
		cmdClean:
		return false
	default:
		return true
	}
}

func CreateBuildPack(arg Arguments, env Environments, config BuildConfig) (bp BuildPack, err error) {
	workDir, err := filepath.Abs(".")
	if err != nil {
		return
	}

	if !common.IsEmptyString(arg.ConfigFile) {
		workDir, _ = filepath.Split(arg.ConfigFile)
	}

	bp.WorkDir = workDir
	bp.Arguments = arg
	bp.Environments = env
	bp.BuildConfig = config
	bp.RepoManager = CreateRepoManager()
	bp.GitManager = CreateGitManager()
	return
}

func (bp *BuildPack) Run(ctx context.Context) error {
	switch bp.Arguments.Command {
	case cmdVersion:
		common.PrintInfo("version %s", version)
		return nil
	case cmdBuild:
		return bp.build()
	case cmdClean:
		return bp.clean()
	case cmdHelp:
		f.Usage()
		return nil
	default:
		f.Usage()
		return nil
	}
}
