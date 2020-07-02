package publisher

type DockerSql struct {
}

func (n DockerSql) PrePublish() error {
	return nil
}

func (n DockerSql) Publish() error {
	return nil
}

func (n DockerSql) PostPublish() error {
	return nil
}

func init() {
	registries["docker_sql"] = &DockerSql{}
}
