package publisher

type ArtifactoryYarn struct {
}

func (n ArtifactoryYarn) PrePublish() error {
	return nil
}

func (n ArtifactoryYarn) Publish() error {
	return nil
}

func (n ArtifactoryYarn) PostPublish() error {
	return nil
}

func init() {
	registries["artifactory_yarn"] = &ArtifactoryYarn{}
}
