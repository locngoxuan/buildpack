package builder

import (
	"errors"
	"fmt"
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack"
	"strings"
	"time"
)

const (
	builderFileName = "builder.yml"
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
	GenerateConfig() error
	LoadConfig(ctx BuildContext) error
	Clean(ctx BuildContext) error
	PreBuild(ctx BuildContext) error
	Build(ctx BuildContext) error
	PostBuild(ctx BuildContext) error
}

func CreateBuilder(bp buildpack.BuildPack, moduleConfig buildpack.ModuleConfig, release bool) (Builder, error) {
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
	versionStr := strings.TrimSpace(b.BuildContext.BuildPack.Config.Version)
	if len(b.BuildContext.BuildPack.RuntimeConfig.Version()) > 0 {
		versionStr = b.BuildContext.BuildPack.RuntimeConfig.Version()
	}

	v, err := buildpack.FromString(versionStr)
	if err != nil {
		return b, err
	}

	if b.BuildContext.Release {
		b.BuildContext.Version = v.WithoutLabel()
	} else {
		label := labelSnapshot
		if len(b.BuildContext.RuntimeConfig.Label()) > 0 {
			t := time.Now()
			buildNumber := t.Format("20060102150405")
			label = fmt.Sprintf("%s.%s", b.BuildContext.RuntimeConfig.Label(), buildNumber)
		}
		b.BuildContext.Version = v.WithLabel(label)
	}
	return b, b.BuildTool.LoadConfig(b.BuildContext)
}

func (b *Builder) ToolName() string {
	return b.BuildTool.Name()
}

func (b *Builder) GenerateConfig() error {
	return nil
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
