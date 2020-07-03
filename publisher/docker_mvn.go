package publisher

type DockerMvn struct {
}

func (n DockerMvn) PrePublish(ctx PublisherContext) error {
	return nil
}

func (n DockerMvn) Publish(ctx PublisherContext) error {
	return nil
}

func (n DockerMvn) PostPublish(ctx PublisherContext) error {
	return nil
}

func init() {
	registries["docker_mvn"] = &DockerMvn{}
}
