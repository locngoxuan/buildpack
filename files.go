package main

import "path/filepath"

func (bp *BuildPack) getBuildPackConfigPath() string {
	p, err := filepath.Abs(filepath.Join(bp.Root, fileBuildPackConfig))
	if err != nil {
		buildError(*bp.Error("", err))
	}
	return p
}

func (bp *BuildPack) getBuilderConfigPath(modulePath string) string {
	p, err := filepath.Abs(filepath.Join(bp.Root, modulePath, fileBuilderConfig))
	if err != nil {
		buildError(*bp.Error("", err))
	}
	return p
}

func (bp *BuildPack) getBuilderSpecificFile(modulePath, filename string) string {
	p, err := filepath.Abs(filepath.Join(bp.Root, modulePath, filename))
	if err != nil {
		buildError(*bp.Error("", err))
	}
	return p
}

func (bp *BuildPack) getModuleWorkingDir(modulePath string) string {
	p, err := filepath.Abs(filepath.Join(bp.Root, modulePath))
	if err != nil {
		buildError(*bp.Error("", err))
	}
	return p
}
