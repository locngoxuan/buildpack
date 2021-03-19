package v1

import (
	"context"
	"errors"
	"github.com/locngoxuan/buildpack/common"
	"path/filepath"
	"sort"
	"strings"
)

func (bp *BuildPack) clean(ctx context.Context) error {
	//create tmp directory
	outputDir := filepath.Join(bp.WorkDir, BuildPackOutputDir)
	err := common.CreateDir(common.CreateDirOption{
		SkipContainer: true,
		AbsPath:       outputDir,
		Perm:          0777,
	})
	if err != nil {
		return err
	}

	ms := make([]Module, 0)
	if common.IsEmptyString(bp.Arguments.Module) {
		for _, module := range bp.BuildConfig.Modules {
			ms = append(ms, Module{
				Id:   module.Id,
				Name: module.Name,
				Path: module.Path,
			})
		}
	} else {
		modules := strings.Split(bp.Arguments.Module, ",")
		mmap := make(map[string]struct{})
		for _, module := range modules {
			mmap[module] = struct{}{}
		}

		for _, module := range bp.BuildConfig.Modules {
			if _, ok := mmap[module.Name]; !ok {
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
		return errors.New("not found any module")
	}

	//sorting by id
	sort.Sort(SortedById(ms))

	for _, module := range bp.BuildConfig.Modules {
		m := Module{
			Id:   module.Id,
			Name: module.Name,
			Path: module.Path,
		}
		common.PrintLog("clean module %s", module.Name)
		err = m.clean(ctx, *bp)
		if err != nil {
			return err
		}
	}

	return nil
}
