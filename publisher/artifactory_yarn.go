package publisher

type ArtifactoryYarn struct {
}

func (n ArtifactoryYarn) PrePublish(ctx PublisherContext) error {
	return nil
}

func (n ArtifactoryYarn) Publish(ctx PublisherContext) error {
	return nil
}

func (n ArtifactoryYarn) PostPublish(ctx PublisherContext) error {
	return nil
}

func init() {
	registries["artifactory_yarn"] = &ArtifactoryYarn{}
}
