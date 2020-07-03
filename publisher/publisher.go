package publisher

import "errors"

var registries = make(map[string]Interface)

const PublishConfigFileName = "Buildpackfile.publish"

type Interface interface {
	PrePublish(ctx PublisherContext) error
	Publish(ctx PublisherContext) error
	PostPublish(ctx PublisherContext) error
}

type DummyPublisher struct {
}

func (n DummyPublisher) PrePublish(ctx PublisherContext) error {
	return nil
}

func (n DummyPublisher) Publish(ctx PublisherContext) error {
	return nil
}

func (n DummyPublisher) PostPublish(ctx PublisherContext) error {
	return nil
}

func GetPublisher(name string) (Interface, error) {
	i, ok := registries[name]
	if !ok {
		return nil, errors.New("not found publisher with name " + name)
	}
	return i, nil
}
