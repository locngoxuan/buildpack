package builder

import (
	"errors"
	"fmt"
	"path/filepath"
	"plugin"
	"strings"
)

const BuildConfigFileName = "Buildpackfile.build"

type Interface interface {
	Clean(ctx BuildContext) error
	PreBuild(ctx BuildContext) error
	Build(ctx BuildContext) error
	PostBuild(ctx BuildContext) error
	PostFail(ctx BuildContext) error
}

func GetBuilder(name string) (Interface, error) {
	if strings.HasPrefix(name, "plugin_") {
		pluginName := fmt.Sprintf("%s.so", name)
		pluginPath := filepath.Join("/etc/buildpack/plugins/builder", pluginName)
		p, err := plugin.Open(pluginPath)
		if err != nil {
			return nil, fmt.Errorf("open %s get error %s", name, err.Error())
		}
		f, err := p.Lookup("GetBuilder")
		if err != nil {
			return nil, fmt.Errorf("find builder from %s get error %s", name, err.Error())
		}
		return f.(func() Interface)(), nil
	}

	switch name {
	case "mvn":
		return &Mvn{}, nil
	case "sql":
		return &Sql{}, nil
	case "sql_lib":
		return &SqlLib{}, nil
	case "sql_app":
		return &SqlApp{}, nil
	case "yarn":
		return &Yarn{}, nil
	default:
		return nil, errors.New("not found builder with name " + name)
	}

	if name == "mvn"{
		return &Mvn{}, nil
	}
	i, ok := registries[name]
	if !ok {
		return nil, errors.New("not found builder with name " + name)
	}
	return i, nil
}
