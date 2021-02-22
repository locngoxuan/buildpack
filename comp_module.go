package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Module struct {
	Id   int
	Name string
	Path string

	moduleDir   string
	output      string
	buildConfig BuildConfig
}

type SortedById []Module

func (a SortedById) Len() int           { return len(a) }
func (a SortedById) Less(i, j int) bool { return a[i].Id < a[j].Id }
func (a SortedById) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func (m *Module) initiate() error {
	m.moduleDir = filepath.Join(workDir, m.Path)
	m.output = filepath.Join(workDir, OutputBuildpack, m.Name)
	err := os.MkdirAll(m.output, 0777)
	if err != nil {
		return err
	}

	m.buildConfig, err = readBuildConfig(m.moduleDir)
	if err != nil {
		return err
	}
	return nil
}

func (m *Module) clean(ctx context.Context) error {
	outputDir := filepath.Join(workDir, OutputBuildpack, m.Name)
	_, err := os.Stat(outputDir)
	if os.IsNotExist(err) {
		return nil
	}
	return os.RemoveAll(outputDir)
}

func (m *Module) build(ctx context.Context) error {
	return nil
}

func (m *Module) pack(ctx context.Context) error {
	return nil
}

func (m *Module) publish(ctx context.Context) error {
	return nil
}

func initModule(id int, name, path string) (Module, error) {
	m := Module{
		Id:   id,
		Name: name,
		Path: path,
	}
	err := m.initiate()
	if err != nil {
		return m, err
	}
	return m, nil
}

//preparing build environment
func prepareListModule() ([]Module, error) {
	ms := make([]Module, 0)
	if isStringEmpty(arg.Module) {
		for _, module := range cfg.Modules {
			m, err := initModule(module.Id, module.Name, module.Path)
			if err != nil {
				return nil, err
			}
			ms = append(ms, m)
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
			m, err := initModule(module.Id, module.Name, module.Path)
			if err != nil {
				return nil, err
			}
			ms = append(ms, m)
		}
	}

	if len(ms) == 0 {
		return nil, errors.New("not found any module")
	}

	//sorting by id
	sort.Sort(SortedById(ms))
	return ms, nil
}
