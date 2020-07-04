package publisher

type ArtifactoryYarn struct {
	ArtifactoryPublisher
}

func (n ArtifactoryYarn) PrePublish(ctx PublishContext) error {
	return nil
}

func (n ArtifactoryYarn) Publish(ctx PublishContext) error {
	return nil
}

func (n ArtifactoryYarn) PostPublish(ctx PublishContext) error {
	return nil
}

func init() {
	yarn := &ArtifactoryYarn{}
	yarn.PreparePackage = func(ctx PublishContext) (packages []ArtifactoryPackage, e error) {
		return nil, nil
	}
	registries["artifactory_yarn"] = yarn
}
