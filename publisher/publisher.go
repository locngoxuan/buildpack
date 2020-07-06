package publisher

import "errors"

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
	i, ok := registries[name]
	if !ok {
		return nil, errors.New("not found publisher with name " + name)
	}
	return i, nil
}
