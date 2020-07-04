package publisher

type DockerYarn struct {
}

func (n DockerYarn) PrePublish(ctx PublishContext) error {
	return nil
}

func (n DockerYarn) Publish(ctx PublishContext) error {
	return nil
}

func (n DockerYarn) PostPublish(ctx PublishContext) error {
	return nil
}

func init() {
	registries["docker_yarn"] = &DockerYarn{}
}
