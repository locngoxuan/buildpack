package main

import (
	"github.com/pkg/errors"
	"strings"
)

type Builder interface {
	CreateContext(bp BuildPack, opt BuildPackModuleRuntimeParams) (BuildContext, error)
	WriteConfig(bp BuildPack, opt BuildPackModuleConfig) error

	Verify(ctx BuildContext) error
	Clean(ctx BuildContext) error
	Build(ctx BuildContext) error
}

type BuildContext struct {
	WorkingDir string
	Name       string
	Path       string
	metadata   map[string]interface{}
	BuildPack
	BuildPackModuleRuntimeParams
}

func (c *BuildContext) Add(key string, value interface{}) {
	c.metadata[key] = value
}

func (c *BuildContext) Get(key string) (interface{}, error) {
	v, ok := c.metadata[key]
	if !ok {
		return nil, errors.New("not found metadata by key " + key)
	}
	return v, nil
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
