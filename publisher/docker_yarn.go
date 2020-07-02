package publisher

type DockerYarn struct {
}

func (n DockerYarn) PrePublish() error {
	return nil
}

func (n DockerYarn) Publish() error {
	return nil
}

func (n DockerYarn) PostPublish() error {
	return nil
}

func init() {
	registries["docker_yarn"] = &DockerYarn{}
}
