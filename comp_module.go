package main

import (
	"context"
	"errors"
	"github.com/locngoxuan/buildpack/v1/common"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Module struct {
	Id   int
	Name string
	Path string
}

type SortedById []Module

func (a SortedById) Len() int           { return len(a) }
func (a SortedById) Less(i, j int) bool { return a[i].Id < a[j].Id }
func (a SortedById) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func (m *Module) clean(ctx context.Context) error {
	outputDir := filepath.Join(workDir, OutputBuildpack, m.Name)
	_, err := os.Stat(outputDir)
	if os.IsNotExist(err) {
		return nil
	}
	return os.RemoveAll(outputDir)
}

func (m *Module) build(ctx context.Context) error {
	var err error

	moduleDir := filepath.Join(workDir, m.Path)
	output := filepath.Join(workDir, OutputBuildpack, m.Name)
	err = os.MkdirAll(output, 0777)
	if err != nil {
		return err
	}

	buildConfig, err := readBuildConfig(moduleDir)
	if err != nil {
		return err
	}

	/**
	- get builder

	- if mvn, yarn, sql --> use internal builder

	- if prefix is custom --> load library
	 */

	if strings.HasPrefix(buildConfig.Builder, "custom") {

	} else {
		switch buildConfig.Builder {
		case "mvn":
			return runMvnBuild(ctx, *m)
		case "yarn":
		case "sql":
		case "-":
			//ignore build process
			return nil
		}
	}
	return err
}

func (m *Module) pack(ctx context.Context) error {
	return nil
}

func (m *Module) publish(ctx context.Context) error {
	return nil
}

func prepareListModule() ([]Module, error) {
	ms := make([]Module, 0)
	if common.IsEmptyString(arg.Module) {
		for _, module := range cfg.Modules {
			ms = append(ms, Module{
				Id:   module.Id,
				Name: module.Name,
				Path: module.Path,
			})
		}
	} else {
		excludes := false
		argModules := strings.Split(arg.Module, ",")
		if strings.HasPrefix(arg.Module, "!") {
			//module exclusion
			argModules = strings.Split(strings.TrimPrefix(arg.Module, "!"), ",")
			excludes = true
		}

		set := make(map[string]struct{})
		for _, module := range argModules {
			set[module] = struct{}{}
		}

		for _, module := range cfg.Modules {
			_, ok := set[module.Name]
			if (excludes && ok) || (!excludes && !ok) {
				continue
			}

			ms = append(ms, Module{
				Id:   module.Id,
				Name: module.Name,
				Path: module.Path,
			})
		}
	}

	if len(ms) == 0 {
		return nil, errors.New("not found any module")
	}

	//sorting by id
	sort.Sort(SortedById(ms))
	return ms, nil
}
