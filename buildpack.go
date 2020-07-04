package buildpack

import (
	"context"
	"errors"
	"io/ioutil"
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack/common"
	"scm.wcs.fortna.com/lngo/buildpack/publisher"
)

const version = "2.0.0"

type BuildPack struct {
	WorkDir string

	Arguments
	Environments
	BuildConfig

	GitManager
}

func (bp BuildPack) IsSkipClean() bool {
	if bp.DevMode {
		return true
	}
	return bp.Arguments.SkipClean
}

func (bp BuildPack) IsSkipContainer() bool {
	if bp.DevMode {
		return true
	}
	return bp.Arguments.SkipContainer
}

func (bp BuildPack) IsSkipPublish() bool {
	if bp.DevMode {
		return true
	}
	return bp.Arguments.SkipPublish
}

func (bp BuildPack) GetVersion() string {
	if common.IsEmptyString(bp.Arguments.Version) {
		return bp.BuildConfig.Version
	}
	return bp.Arguments.Version
}

func (bp BuildPack) validateDocker() error {
	if !bp.IsSkipContainer() {
		return common.ValidateDockerHostConnection()
	}
	return nil
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

func createRepoManager(c BuildConfig) (rm publisher.RepoManager, err error) {
	rm.Repos = make(map[string]publisher.Repository)
	if c.Repos == nil || len(c.Repos) == 0 {
		err = errors.New("not found any repository configuration")
		return
	}
	for _, repo := range c.Repos {
		r := publisher.Repository{
			Name: repo.Name,
		}
		if repo.Stable != nil {
			r.Stable = &publisher.RepoChannel{
				Address:  repo.Stable.Address,
				Username: repo.Stable.Username,
				Password: repo.Stable.Password,
			}
		}
		if repo.Unstable != nil {
			r.Unstable = &publisher.RepoChannel{
				Address:  repo.Unstable.Address,
				Username: repo.Unstable.Username,
				Password: repo.Unstable.Password,
			}
		}
		rm.Repos[r.Name] = r
	}
	return
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

	//if skip publish then no need to create repo manager
	if !bp.SkipPublish {
		rm, e := createRepoManager(config)
		if e != nil {
			err = e
			return
		}
		publisher.SetRepoManager(rm)
	}

	//if skip git is true then no need to create git manager
	if !bp.SkipGit {
		bp.GitManager = CreateGitManager(config)
	}

	if bp.BuildConfig.Docker != nil {
		common.SetDockerHost(bp.BuildConfig.Docker.Hosts)
	}
	err = bp.validateDocker()
	return
}

func (bp *BuildPack) Run(ctx context.Context) error {
	switch bp.Arguments.Command {
	case cmdVersion:
		common.PrintInfo("version %s", version)
		return nil
	case cmdBuild:
		common.SetLogOutput(ioutil.Discard)
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
