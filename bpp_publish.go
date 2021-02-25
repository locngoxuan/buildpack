package main

import (
	"context"
	"fmt"
	"github.com/locngoxuan/buildpack/config"
	"github.com/locngoxuan/buildpack/core"
	"github.com/locngoxuan/buildpack/instrument"
	"github.com/locngoxuan/buildpack/utils"
)

func publish(ctx context.Context) error {
	//preparing phase of build process is started
	if utils.IsNotExists(outputDir) {
		return fmt.Errorf("output directory %s does not exist", config.OutputDir)
	}

	buildInfo, err := config.ReadBuildOutputInfo(outputDir)
	if err != nil {
		return err
	}

	if utils.IsStringEmpty(buildInfo.Version) {
		return fmt.Errorf("not found build info")
	}

	modules, err := prepareListModule()
	if err != nil {
		return err
	}

	if len(modules) == 0 {
		return fmt.Errorf("could not find the selected module")
	}

	projectRepoConfig, err := core.ReadProjectRepositoryConfig(workDir, arg.ConfigFile)
	if err != nil {
		return err
	}

	globalRepoConfig, err := core.ReadGlobalRepositoryConfig()
	if err != nil {
		return err
	}

	repositories := make(map[string]config.Repository)
	for _, r := range globalRepoConfig.Repos {
		repositories[r.Id] = r
	}
	for _, r := range projectRepoConfig.Repos {
		repositories[r.Id] = r
	}

	for _, module := range modules {
		for _, pc := range module.config.Publish {
			if len(pc.RepoIds) == 0 {
				continue
			}

			selectedRepos := make(map[string]config.Repository)
			for _, repoId := range pc.RepoIds {
				r, ok := repositories[repoId]
				if !ok {
					continue
				}
				selectedRepos[repoId] = r
			}
			resp := instrument.PublishPackage(ctx, instrument.PublishRequest{
				BaseProperties: instrument.BaseProperties{
					WorkDir:       workDir,
					OutputDir:     outputDir,
					ShareDataDir:  arg.ShareData,
					Release:       buildInfo.Release,
					Patch:         buildInfo.Release,
					Version:       buildInfo.Version,
					ModulePath:    module.Path,
					ModuleName:    module.Name,
					ModuleOutputs: module.config.Output,
					LocalBuild:    arg.BuildLocal,
				},
				Repositories: selectedRepos,
				PublishConfig: config.PublishConfig{
					Type:    pc.Type,
					RepoIds: pc.RepoIds,
				},
			})
			if resp.Err != nil {
				if resp.ErrStack != "" {
					return fmtError(resp.Err, resp.ErrStack)
				}
				return resp.Err
			}
		}
	}
	return nil
}
