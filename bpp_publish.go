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

	repositories := make(map[string]core.Repository)
	for _, r := range globalRepoConfig.Repos {
		repositories[r.Id] = r
	}
	for _, r := range projectRepoConfig.Repos {
		repositories[r.Id] = r
	}
	for _, module := range modules {
		fmt.Println(module)
		resp := instrument.PublishPackage(ctx, instrument.PublishRequest{
			Repositories: repositories,
		})
		if resp.Err != nil {
			if resp.ErrStack != "" {
				return fmtError(resp.Err, resp.ErrStack)
			}
			return resp.Err
		}
	}
	return nil
}
