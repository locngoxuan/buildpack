package instrument

import "context"

const ArtifactoryYarnPublisherName = "artifactoryyarn"

func publishYarnJarToArtifactory(ctx context.Context, request PublishRequest) Response {
	return responseSuccess()
}
