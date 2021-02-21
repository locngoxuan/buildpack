package main

import (
	"context"
	"errors"
	"github.com/locngoxuan/buildpack/v1/common"
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
	return nil
}

func (m *Module) build() {

}

func (m *Module) pack() {

}

func (m *Module) publish() {

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
