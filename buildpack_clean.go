package buildpack

import (
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack/common"
)

func (bp *BuildPack) clean() error {
	//create tmp directory
	outputDir := filepath.Join(bp.WorkDir, BuildPackOutputDir)
	err := common.DeleteDir(outputDir, true)
	if err != nil {
		return err
	}
	for _, module := range bp.BuildConfig.Modules {
		m := Module{
			Id:   module.Id,
			Name: module.Name,
			Path: module.Path,
		}
		err = m.clean(*bp)
		if err != nil {
			return err
		}
	}
	return nil
}
