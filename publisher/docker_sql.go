package publisher

type DockerSql struct {
}

func (n DockerSql) PrePublish(ctx PublisherContext) error {
	return nil
}

func (n DockerSql) Publish(ctx PublisherContext) error {
	return nil
}

func (n DockerSql) PostPublish(ctx PublisherContext) error {
	return nil
}

func init() {
	registries["docker_sql"] = &DockerSql{}
}
