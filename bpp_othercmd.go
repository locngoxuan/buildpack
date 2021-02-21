package main

import (
	"context"
	"fmt"
)

func showVersion() error  {
	fmt.Printf("version: %s\n", version)
	return nil
}

func clean(ctx context.Context) error{
	var err error
	cfg, err = ReadConfig(arg.ConfigFile)
	if err != nil{
		return nil
	}
	modules, err := prepareListModule()
	if err != nil{
		return err
	}

	for _, module := range modules{
		err = module.clean(ctx)
		if err != nil{
			return err
		}
	}
	return nil
}
