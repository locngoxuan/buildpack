package builder

import (
	"github.com/pkg/errors"
	. "scm.wcs.fortna.com/lngo/buildpack"
)

type Builder interface {
	CreateContext(bp *BuildPack, opt ModuleRuntime) (BuildContext, error)
	WriteConfig(bp BuildPack, opt ModuleConfig) error

	Verify(ctx BuildContext) error
	Clean(ctx BuildContext) error
	UnitTest(ctx BuildContext) error
	Build(ctx BuildContext) error
}

type BuildContext struct {
	WorkingDir string
	Name       string
	Path       string
	metadata   map[string]interface{}
	*BuildPack
	ModuleRuntime
}

func NewBuildContext(workingDir, name, path string) BuildContext {
	return BuildContext{
		WorkingDir: workingDir,
		Name:       name,
		Path:       path,
		metadata:   make(map[string]interface{}),
	}
}

func (c *BuildContext) Add(key string, value interface{}) {
	c.metadata[key] = value
}

func (c *BuildContext) Get(key string) (interface{}, error) {
	v, ok := c.metadata[key]
	if !ok {
		return nil, errors.New("not found metadata by key " + key)
	}
	return v, nil
}

var builders map[string]Builder

func init() {
	builders = make(map[string]Builder)
	builders[buildTypeMvn] = &MVN{}
}

func VerifyBuilder(name string) bool {
	_, exist := builders[name]
	return exist
}

func GetBuilder(builderName string) (Builder, error) {
	builder, ok := builders[builderName]
	if !ok {
		return nil, errors.New("can not find builder by name: " + builderName)
	}
	return builder, nil
}
