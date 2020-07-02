package publisher

type ArtifactorySql struct {
}

func (n ArtifactorySql) PrePublish() error {
	return nil
}

func (n ArtifactorySql) Publish() error {
	return nil
}

func (n ArtifactorySql) PostPublish() error {
	return nil
}

func init() {
	registries["artifactory_sql"] = &ArtifactorySql{}
}
