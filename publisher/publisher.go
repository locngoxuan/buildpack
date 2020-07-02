package publisher

var registries = make(map[string]Interface)

const PublishConfigFileName = "Buildpackconfig.publish"

type Interface interface {
	PrePublish() error
	Publish() error
	PostPublish() error
}

type DummyPublisher struct {
}

func (n DummyPublisher) PrePublish() error {
	return nil
}

func (n DummyPublisher) Publish() error {
	return nil
}

func (n DummyPublisher) PostPublish() error {
	return nil
}
