package buildpack

import (
	"context"
	"sort"
	"strings"
)

const version = "2.0.0"

type BuildPack struct {
	Arguments
	Environments
	BuildConfig

	RepoManager
	GitManager
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

func CreateBuildPack(arg Arguments, env Environments, config BuildConfig) BuildPack {
	rm := CreateRepoManager()
	gm := CreateGitManager()
	return BuildPack{
		arg,
		env,
		config,
		rm,
		gm,
	}
}

func (bp *BuildPack) Run(ctx context.Context) error {
	switch bp.Arguments.Command {
	case cmdVersion:
		PrintInfo("version %s", version)
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

func (bp *BuildPack) handle(){
	modules := strings.Split(bp.Arguments.Module, ",")
	ms := make([]Module, 0)
	if len(modules) == 0 {
		for _, module := range bp.BuildConfig.Modules {
			ms = append(ms, Module{
				Id:   module.Id,
				Name: module.Name,
				Path: module.Path,
			})
		}
	} else {
		mmap := make(map[string]struct{})
		for _, module := range modules {
			mmap[module] = struct{}{}
		}

		for _, module := range bp.BuildConfig.Modules {
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

	sort.Sort(SortedById(ms))

	for _, module := range ms {
		module.start()
	}
}
