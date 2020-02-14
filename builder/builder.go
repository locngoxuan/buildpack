package builder

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack"
)

type Builder struct {
	BuildTool
	BuildContext
}

type BuildContext struct {
	Name string
	Path string
	buildpack.BuildPack
	Release    bool
	Version    string
	WorkingDir string
	Values     map[string]interface{}
}

func (bc *BuildContext) GetFile(args ...string) string {
	parts := []string{
		bc.WorkingDir,
	}
	parts = append(parts, args...)
	p, err := filepath.Abs(filepath.Join(parts...))
	if err != nil {
		buildpack.LogFatal(bc.Error("", err))
	}
	return p
}

type BuildTool interface {
	Name() string
	GenerateConfig(ctx BuildContext) error
	LoadConfig(ctx BuildContext) error
	Clean(ctx BuildContext) error
	PreBuild(ctx BuildContext) error
	Build(ctx BuildContext) error
	PostBuild(ctx BuildContext) error
}

func CreateBuilder(bp buildpack.BuildPack, moduleConfig buildpack.ModuleConfig, release bool, version string) (Builder, error) {
	b := Builder{
		BuildContext: BuildContext{
			moduleConfig.Name,
			moduleConfig.Path,
			bp,
			release,
			bp.Config.Version,
			bp.GetModuleWorkingDir(moduleConfig.Path),
			make(map[string]interface{}),
		},
	}

	tool, ok := buildTools[moduleConfig.BuildTool]
	if !ok {
		return b, errors.New("not found build tool associated to name " + moduleConfig.BuildTool)
	}

	b.BuildTool = tool
	b.BuildContext.Version = version
	return b, b.BuildTool.LoadConfig(b.BuildContext)
}

func (b *Builder) ToolName() string {
	return b.BuildTool.Name()
}

func (b *Builder) GenerateConfig() error {
	return b.BuildTool.GenerateConfig(b.BuildContext)
}

func (b *Builder) Clean() error {
	return b.BuildTool.Clean(b.BuildContext)
}

func (b *Builder) PreBuild() error {
	return b.BuildTool.PreBuild(b.BuildContext)
}

func (b *Builder) Build() error {
	return b.BuildTool.Build(b.BuildContext)
}

func (b *Builder) PostBuild() error {
	return b.BuildTool.PostBuild(b.BuildContext)
}

func copyDirectory(bp buildpack.BuildPack, scrDir, dest string) error {
	entries, err := ioutil.ReadDir(scrDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(scrDir, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			return err
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := createIfNotExists(destPath, 0755); err != nil {
				return err
			}
			if err := copyDirectory(bp, sourcePath, destPath); err != nil {
				return err
			}
		default:
			_, name := filepath.Split(sourcePath)
			buildpack.LogVerbose(bp, fmt.Sprintf("Copying %s to %s", name, destPath))
			if err := copy(sourcePath, destPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func copy(srcFile, dstFile string) error {
	out, err := os.Create(dstFile)
	if err != nil {
		return err
	}

	defer func() {
		_ = out.Close()
	}()

	in, err := os.Open(srcFile)
	defer func() {
		_ = in.Close()
	}()
	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return nil
}

func exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	return true
}

func createIfNotExists(dir string, perm os.FileMode) error {
	if exists(dir) {
		return nil
	}

	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("failed to create directory: '%s', error: '%s'", dir, err.Error())
	}

	return nil
}