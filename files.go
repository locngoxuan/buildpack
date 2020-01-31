package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func (bp *BuildPack) getPublishDirectory() string {
	p, err := filepath.Abs(filepath.Join(bp.Root, publishDir))
	if err != nil {
		buildError(*bp.Error("", err))
	}
	return p
}

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

func (bp *BuildPack) buildPathOnRoot(args ...string) string {
	parts := []string{
		bp.Root,
	}
	parts = append(parts, args...)
	p, err := filepath.Abs(filepath.Join(parts...))
	if err != nil {
		buildError(*bp.Error("", err))
	}
	return p
}

func copyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		_ = source.Close()
	}()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		_ = destination.Close()
	}()
	_, err = io.Copy(destination, source)
	return err
}
