package main

import (
	"context"
	"fmt"
	"github.com/locngoxuan/buildpack/config"
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

	tempModules, err := prepareListModule()
	if err != nil {
		return err
	}

	if len(tempModules) == 0 {
		return fmt.Errorf("could not find the selected module")
	}

	//ignore module that is not configured for building
	modules := make([]Module, 0)
	for _, m := range tempModules {
		if len(m.config.Publish) == 0 {
			continue
		}
		modules = append(modules, m)
	}

	//verify publish type
	for _, m := range modules {
		for _, p := range m.config.Publish {
			if utils.IsStringEmpty(p.Type) {
				return fmt.Errorf("publish type of module %s is malformed", m.Name)
			}
		}
	}

	globalRepoConfig, err := config.ReadGlobalRepositoryConfig()
	if err != nil {
		return err
	}

	repositories := make(map[string]config.Repository)
	for _, r := range globalRepoConfig.Repos {
		repositories[r.Id] = r
	}
	for _, r := range cfg.RepoConfig {
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
					DevMode:       !buildInfo.Release,
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
