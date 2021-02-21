package v1

import (
	"context"
	"errors"
	"fmt"
	"github.com/locngoxuan/buildpack/v1/common"
	"github.com/locngoxuan/buildpack/v1/publisher"
	"os"
	"path/filepath"
	"strings"
)

var Version = "2.0.0"

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
		cmdClean,
		cmdPump:
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

func createRepoManager(workDir string, arg Arguments, c BuildConfig) (rm publisher.RepoManager, err error) {
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

	//validate credential of repositories
	ms := make([]Module, 0)
	if common.IsEmptyString(arg.Module) {
		for _, module := range c.Modules {
			ms = append(ms, Module{
				Id:   module.Id,
				Name: module.Name,
				Path: module.Path,
			})
		}
	} else {
		modules := strings.Split(arg.Module, ",")
		mmap := make(map[string]struct{})
		for _, module := range modules {
			mmap[module] = struct{}{}
		}

		for _, module := range c.Modules {
			if _, ok := mmap[module.Name]; !ok {
				continue
			}
			ms = append(ms, Module{
				Id:   module.Id,
				Name: module.Name,
				Path: module.Path,
			})
		}
	}

	if len(ms) == 0 {
		return
	}

	for _, m := range ms {
		c, e := publisher.ReadConfig(filepath.Join(workDir, m.Path))
		if e != nil {
			err = e
			break
		}
		if c.Publisher == "none" || c.Publisher == "no_publisher" || c.Publisher == "" {
			continue
		}

		repo, ok := rm.Repos[c.Repository]
		if !ok {
			err = fmt.Errorf("repo %s may be not registered", c.Repository)
			break
		}

		if repo.Stable != nil && !repo.Stable.NoAuth {
			if common.IsEmptyString(repo.Stable.Username) || common.IsEmptyString(repo.Stable.Password) {
				err = fmt.Errorf("missing credential of stable channel of repo %s", repo.Name)
				break
			}
		}

		if repo.Unstable != nil && !repo.Unstable.NoAuth {
			if common.IsEmptyString(repo.Unstable.Username) || common.IsEmptyString(repo.Unstable.Password) {
				err = fmt.Errorf("missing credential of unstable channel of repo %s", repo.Name)
				break
			}
		}
	}
	return
}

func CreateBuildPack(arg Arguments, config BuildConfig) (BuildPack, error) {
	bp := BuildPack{}
	var err error
	workDir, err := filepath.Abs(".")
	if err != nil {
		return bp, err
	}

	if !common.IsEmptyString(arg.ConfigFile) {
		workDir, _ = filepath.Split(arg.ConfigFile)
	}

	bp.WorkDir = workDir
	bp.Arguments = arg
	bp.BuildConfig = config

	if bp.Command == cmdPump {
		//requires git configuration
		mainBranch := "master"
		if strings.TrimSpace(config.Git.MainBranch) != "" {
			mainBranch = strings.TrimSpace(config.Git.MainBranch)
		}

		cli, err := createGitManager(workDir, config)
		if err != nil {
			return bp, err
		}
		err = cli.OpenCurrentRepo(mainBranch)
		if err != nil {
			return bp, err
		}
		common.SetGitClient(cli)
	} else if bp.Command == cmdBuild && !bp.IsSkipPublish() {
		rm, err := createRepoManager(workDir, arg, config)
		if err != nil {
			return bp, err
		}
		publisher.SetRepoManager(rm)
		if bp.BuildConfig.Docker != nil {
			common.SetDockerHost(bp.BuildConfig.Docker.Hosts)
		}
		err = bp.validateDocker()
		if err != nil {
			return bp, err
		}
	}
	return bp, nil
}

func (bp *BuildPack) Run(ctx context.Context) error {
	switch bp.Arguments.Command {
	case cmdVersion:
		common.PrintLog("version %s", Version)
		return nil
	case cmdClean:
		defer clearOnExit(ctx, bp)
		return bp.clean(ctx)
	case cmdPump:
		return bp.pump(ctx)
	case cmdBuild:
		defer clearOnExit(ctx, bp)
		return bp.build(ctx)
	case cmdHelp:
		f.Usage()
		return nil
	default:
		f.Usage()
		return nil
	}
}

func clearOnExit(ctx context.Context, bp *BuildPack) {
	if ctx.Err() == nil && !bp.IsSkipClean() {
		outputDir := filepath.Join(bp.WorkDir, BuildPackOutputDir)
		_ = common.DeleteDir(common.DeleteDirOption{
			AbsPath:       outputDir,
			SkipContainer: true,
		})
	}
}

func (bp *BuildPack) Exist(ctx context.Context) {
	if ctx.Err() == nil && !bp.IsSkipClean() {
		outputDir := filepath.Join(bp.WorkDir, BuildPackOutputDir)
		_ = common.DeleteDir(common.DeleteDirOption{
			AbsPath:       outputDir,
			SkipContainer: true,
		})
	}
}
