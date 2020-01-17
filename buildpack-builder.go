package main

import "github.com/pkg/errors"

var builders map[string]Builder

func init() {
	builders = make(map[string]Builder)
}

func getBuilder(builderName string) (Builder, error) {
	builder, ok := builders[builderName]
	if !ok {
		return nil, errors.New("can not find builder by name " + builderName)
	}
	return builder, nil
}
