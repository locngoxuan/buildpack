package publisher

import (
	"errors"
	"fmt"
	"path/filepath"
	"plugin"
	"strings"
)

const PublishConfigFileName = "Buildpackfile.publish"

type Interface interface {
	PrePublish(ctx PublishContext) error
	Publish(ctx PublishContext) error
	PostPublish(ctx PublishContext) error
}

type DummyPublisher struct {
}

func (n DummyPublisher) PrePublish(ctx PublishContext) error {
	return nil
}

func (n DummyPublisher) Publish(ctx PublishContext) error {
	return nil
}

func (n DummyPublisher) PostPublish(ctx PublishContext) error {
	return nil
}

var noPublisher = &DummyPublisher{}

func GetPublisher(name string) (Interface, error) {
	if strings.HasPrefix(name, "plugin.") {
		pluginName := strings.TrimPrefix(name, "plugin.")
		parts := strings.Split(pluginName, ".")
		pluginName = fmt.Sprintf("%s.so", parts[0])
		pluginPath := filepath.Join("/etc/buildpack/plugins/publisher", pluginName)
		p, err := plugin.Open(pluginPath)
		if err != nil {
			return nil, err
		}
		funcName := "GetPublisher"
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
	case "no_publisher",
		"none":
		return noPublisher, nil
	case "artifactory_mvn":
		return getArtifactoryMvn(), nil
	case "artifactory_sql":
		return getArtifactorySql(), nil
	case "docker_sql":
		return getDockerSql(), nil
	default:
		return nil, errors.New("not found publisher with name " + name)
	}
}
