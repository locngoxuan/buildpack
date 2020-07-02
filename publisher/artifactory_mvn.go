package publisher

type ArtifactoryMvn struct {
}

func (n ArtifactoryMvn) PrePublish() error {
	return nil
}

func (n ArtifactoryMvn) Publish() error {
	return nil
}

func (n ArtifactoryMvn) PostPublish() error {
	return nil
}

func init() {
	registries["artifactory_mvn"] = &ArtifactoryMvn{}
}
