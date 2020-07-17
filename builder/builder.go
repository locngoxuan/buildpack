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
	if strings.HasPrefix(name, "plugin.") {
		pluginName := strings.TrimPrefix(name, "plugin.")
		parts := strings.Split(pluginName, ".")
		pluginName = fmt.Sprintf("%s.so", parts[0])
		pluginPath := filepath.Join("/etc/buildpack/plugins/builder", pluginName)
		p, err := plugin.Open(pluginPath)
		if err != nil {
			return nil, err
		}
		funcName := "GetBuilder"
		if len(parts) > 1 {
			funcName = parts[1]
		}
		f, err := p.Lookup(funcName)
		if err != nil {
			return nil, err
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
}
