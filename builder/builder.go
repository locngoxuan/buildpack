package builder

import (
	"errors"
	"io"
	"os"
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

var logOutput io.Writer = os.Stdout

func SetOutput(w io.Writer){
	logOutput = w
}

func GetBuilder(name string) (Interface, error) {
	i, ok := registries[name]
	if !ok {
		return nil, errors.New("not found builder with name " + name)
	}
	return i, nil
}
