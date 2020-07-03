package publisher

type DockerYarn struct {
}

func (n DockerYarn) PrePublish(ctx PublisherContext) error {
	return nil
}

func (n DockerYarn) Publish(ctx PublisherContext) error {
	return nil
}

func (n DockerYarn) PostPublish(ctx PublisherContext) error {
	return nil
}

func init() {
	registries["docker_yarn"] = &DockerYarn{}
}
