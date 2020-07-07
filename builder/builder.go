package builder

import (
	"errors"
)

var registries = make(map[string]Interface)

const BuildConfigFileName = "Buildpackfile.build"

type Interface interface {
	Clean(ctx BuildContext) error
	PreBuild(ctx BuildContext) error
	Build(ctx BuildContext) error
	PostBuild(ctx BuildContext) error
	PostFail(ctx BuildContext) error
}

func GetBuilder(name string) (Interface, error) {
	i, ok := registries[name]
	if !ok {
		return nil, errors.New("not found builder with name " + name)
	}
	return i, nil
}
