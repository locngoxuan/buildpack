package publisher

type DockerMvn struct {
}

func (n DockerMvn) PrePublish() error {
	return nil
}

func (n DockerMvn) Publish() error {
	return nil
}

func (n DockerMvn) PostPublish() error {
	return nil
}

func init() {
	registries["docker_mvn"] = &DockerMvn{}
}
