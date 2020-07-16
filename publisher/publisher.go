package publisher

import (
	"errors"
	"fmt"
	"path/filepath"
	"plugin"
	"strings"
)

var registries = make(map[string]Interface)

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

func init() {
	registries["no_publisher"] = &DummyPublisher{}
}

func GetPublisher(name string) (Interface, error) {
	if strings.HasPrefix(name, "plugin_") {
		pluginName := fmt.Sprintf("%s.so", name)
		pluginPath := filepath.Join("/etc/buildpack/plugins/publisher", pluginName)
		p, err := plugin.Open(pluginPath)
		if err != nil {
			return nil, fmt.Errorf("open %s get error %s", name, err.Error())
		}
		f, err := p.Lookup("GetPublisher")
		if err != nil {
			return nil, fmt.Errorf("find builder from %s get error %s", name, err.Error())
		}
		return f.(func() Interface)(), nil
	}
	i, ok := registries[name]
	if !ok {
		return nil, errors.New("not found publisher with name " + name)
	}
	return i, nil
}
