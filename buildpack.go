package buildpack

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack/common"
	"scm.wcs.fortna.com/lngo/buildpack/publisher"
	"strings"
)

const version = "2.0.0"

type BuildPack struct {
	WorkDir string
	Arguments
	BuildConfig
}

func (bp BuildPack) IsSkipClean() bool {
	if bp.DevMode {
		return true
	}
	return bp.Arguments.SkipClean
}

func (bp BuildPack) GetConfigFile() string {
	if !common.IsEmptyString(bp.Arguments.ConfigFile) {
		return bp.Arguments.ConfigFile
	}
	return filepath.Join(bp.WorkDir, ConfigFileName)
}

func (bp BuildPack) IsSkipContainer() bool {
	if bp.DevMode {
		return true
	}
	return bp.Arguments.SkipContainer
}

func (bp BuildPack) IsSkipPublish() bool {
	if bp.DevMode || bp.Arguments.Command == cmdClean {
		return true
	}
	return bp.Arguments.SkipPublish
}

func (bp BuildPack) IsSkipGit() bool {
	if bp.DevMode || (!bp.BuildRelease && !bp.BuildPath) || bp.Arguments.Command == cmdClean {
		return true
	}
	return bp.Arguments.SkipGit
}

func (bp BuildPack) IsSkipGitBraching() bool {
	if bp.IsSkipGit() || bp.DevMode || bp.BuildPath {
		return true
	}
	return bp.Arguments.SkipBranching
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

func createGitManager(root string, c BuildConfig) (cli common.GitClient, err error) {
	if c.Git == nil {
		err = errors.New("not found git configuration")
		return
	}
	cli.WorkDir = root
	if common.IsEmptyString(c.Git.AccessToken) {
		err = errors.New("access token must not be empty")
		return
	}
	cli.Name = c.Git.Username
	if common.IsEmptyString(c.Git.Username) {
		cli.Name = "Build System"
	}

	cli.Email = c.Git.Email
	if common.IsEmptyString(c.Git.Email) {
		cli.Email = "xuanloc0511@gmail.com"
	}

	token := strings.TrimSpace(c.Git.AccessToken)
	if strings.HasPrefix(token, "$") {
		token = os.ExpandEnv(token)
	}
	cli.AccessToken = token
	return
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
			user := strings.TrimSpace(repo.Stable.Username)
			if strings.HasPrefix(user, "$") {
				user = os.ExpandEnv(user)
			}
			password := strings.TrimSpace(repo.Stable.Password)
			if strings.HasPrefix(password, "$") {
				password = os.ExpandEnv(password)
			}

			r.Stable = &publisher.RepoChannel{
				Address:  repo.Stable.Address,
				NoAuth:   repo.Stable.NoAuth,
				Username: user,
				Password: password,
			}
		}
		if repo.Unstable != nil {
			user := strings.TrimSpace(repo.Unstable.Username)
			if strings.HasPrefix(user, "$") {
				user = os.ExpandEnv(user)
			}
			password := strings.TrimSpace(repo.Unstable.Password)
			if strings.HasPrefix(password, "$") {
				password = os.ExpandEnv(password)
			}

			r.Unstable = &publisher.RepoChannel{
				Address:  repo.Unstable.Address,
				NoAuth:   repo.Stable.NoAuth,
				Username: user,
				Password: password,
			}
		}
		rm.Repos[r.Name] = r
	}

	for _, repo := range rm.Repos {
		if repo.Stable != nil && !repo.Stable.NoAuth {
			if common.IsEmptyString(repo.Stable.Username) || common.IsEmptyString(repo.Stable.Password) {
				err = fmt.Errorf("missing credential of stable channel of repo %s", repo.Name)
				return
			}
		}

		if repo.Unstable != nil && !repo.Unstable.NoAuth {
			if common.IsEmptyString(repo.Unstable.Username) || common.IsEmptyString(repo.Unstable.Password) {
				err = fmt.Errorf("missing credential of unstable channel of repo %s", repo.Name)
				return
			}
		}
	}
	return
}

func CreateBuildPack(arg Arguments, config BuildConfig) (bp BuildPack, err error) {
	workDir, err := filepath.Abs(".")
	if err != nil {
		return
	}

	if !common.IsEmptyString(arg.ConfigFile) {
		workDir, _ = filepath.Split(arg.ConfigFile)
	}

	bp.WorkDir = workDir
	bp.Arguments = arg
	bp.BuildConfig = config

	//if skip publish then no need to create repo manager
	if !bp.IsSkipPublish() {
		rm, e := createRepoManager(config)
		if e != nil {
			err = e
			return
		}
		publisher.SetRepoManager(rm)
	}

	//if skip git is true then no need to create git manager
	if !bp.IsSkipGit() {
		cli, e := createGitManager(workDir, config)
		if err != nil {
			err = e
			return
		}
		err = cli.OpenCurrentRepo()
		if err != nil {
			err = e
			return
		}
		common.SetGitClient(cli)
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
		common.PrintLog("version %s", version)
		return nil
	case cmdBuild:
		defer func() {
			if ctx.Err() == nil && !bp.IsSkipClean() {
				outputDir := filepath.Join(bp.WorkDir, BuildPackOutputDir)
				_ = common.DeleteDir(common.DeleteDirOption{
					AbsPath:       outputDir,
					SkipContainer: true,
				})
			}
		}()
		return bp.build(ctx)
	case cmdClean:
		defer func() {
			outputDir := filepath.Join(bp.WorkDir, BuildPackOutputDir)
			_ = common.DeleteDir(common.DeleteDirOption{
				AbsPath:       outputDir,
				SkipContainer: true,
			})
		}()
		return bp.clean(ctx)
	case cmdHelp:
		f.Usage()
		return nil
	default:
		f.Usage()
		return nil
	}
}

func (bp *BuildPack) Exist(ctx context.Context) {
}
