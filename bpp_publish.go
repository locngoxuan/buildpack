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

	modules, err := prepareListModule()
	if err != nil {
		return err
	}

	for _, module := range modules {
		fmt.Println(module)
		resp := instrument.PublishPackage(ctx, instrument.PublishRequest{})
		if resp.Err != nil {
			return err
		}
	}
	return nil
}
