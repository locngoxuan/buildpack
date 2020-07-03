package buildpack

import (
	"errors"
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack/common"
	"sort"
	"strings"
)

const BuildPackOutputDir = ".buildpack"

func (bp *BuildPack) build() error {
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

	//create tmp directory
	outputDir := filepath.Join(bp.WorkDir, BuildPackOutputDir)
	err := common.DeleteDir(outputDir, true)
	if err != nil {
		return err
	}

	err = common.CreateDir(outputDir, true, 0755)
	if err != nil {
		return err
	}
	for _, module := range ms {
		err := common.CreateDir(filepath.Join(outputDir, module.Name), true, 0755)
		if err != nil {
			return err
		}
	}

	//build
	for _, module := range ms {
		err = module.start(*bp)
		if err != nil {
			return err
		}
	}
	return nil
}
