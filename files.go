package buildpack

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func (bp *BuildPack) GetPublishDirectory() string {
	p, err := filepath.Abs(filepath.Join(bp.Root, PublishDirectory))
	if err != nil {
		LogFatal(*bp.Error("", err))
	}
	return p
}

func (bp *BuildPack) GetBuildPackConfigPath() string {
	p, err := filepath.Abs(filepath.Join(bp.Root, FileBuildPackConfig))
	if err != nil {
		LogFatal(*bp.Error("", err))
	}
	return p
}

func (bp *BuildPack) GetBuilderConfigPath(modulePath string) string {
	p, err := filepath.Abs(filepath.Join(bp.Root, modulePath, FileBuilderConfig))
	if err != nil {
		LogFatal(*bp.Error("", err))
	}
	return p
}

func (bp *BuildPack) GetBuilderSpecificFile(modulePath, filename string) string {
	p, err := filepath.Abs(filepath.Join(bp.Root, modulePath, filename))
	if err != nil {
		LogFatal(*bp.Error("", err))
	}
	return p
}

func (bp *BuildPack) GetModuleWorkingDir(modulePath string) string {
	p, err := filepath.Abs(filepath.Join(bp.Root, modulePath))
	if err != nil {
		LogFatal(*bp.Error("", err))
	}
	return p
}

func (bp *BuildPack) BuildPathOnRoot(args ...string) string {
	parts := []string{
		bp.Root,
	}
	parts = append(parts, args...)
	p, err := filepath.Abs(filepath.Join(parts...))
	if err != nil {
		LogFatal(*bp.Error("", err))
	}
	return p
}

func CopyFile(src, dst string) error {
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
