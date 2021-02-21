package main

import (
	"context"
	"os"
	"path/filepath"
)

func build(ctx context.Context) error {
	var err error
	cfg, err = readProjectConfig(arg.ConfigFile)
	if err != nil {
		return nil
	}
	modules, err := prepareListModule()
	if err != nil {
		return err
	}

	//create .buildpack directory
	output := filepath.Join(workDir, OutputBuildpack)
	if !isNotExists(output){
		err = os.RemoveAll(output)
		if err != nil{
			return err
		}
	}
	err = os.Mkdir(output, 0777)
	if err != nil{
		return err
	}
	//defer func(p string) {
	//	_ = os.RemoveAll(p)
	//}(output)

	for _, module := range modules {
		err = module.build(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
