package main

import (
	"github.com/pkg/errors"
	"strings"
)

type Builder interface {
	LoadConfig() error
	Clean() error
	Build() error
}

var builders map[string]Builder

func init() {
	builders = make(map[string]Builder)
	builders["mvn"] = &BuilderMvn{}
	builders["make"] = nil
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
		return nil, errors.New("can not find builder by name " + builderName)
	}
	return builder, nil
}
