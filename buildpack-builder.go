package main

import (
	"github.com/pkg/errors"
	"strings"
)

type Builder interface {
	LoadConfig(rtOpt BuildPackModuleRuntimeParams, bp BuildPack) error
	WriteConfig(name, path string, opt BuildPackModuleConfig) error
	Clean() error
	Build() error

	SetBuilderPack(bp BuildPack)
}

var builders map[string]Builder

func init() {
	builders = make(map[string]Builder)
	builders[builderTypeMvn] = &BuilderMvn{}
}

func builderOptions() string {
	names := make([]string, 0)
	for name, _ := range builders {
		names = append(names, name)
	}
	return strings.Join(names, "/")
}

func getBuilder(builderName string) (Builder, error) {
	builder, ok := builders[builderName]
	if !ok {
		return nil, errors.New("can not find builder by name: " + builderName)
	}
	return builder, nil
}
