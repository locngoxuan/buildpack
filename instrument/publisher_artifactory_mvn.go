package instrument

import "context"

const ArtifactoryMvnPublisherName = "artifactorymvn"

func publishMvnJarToArtifactory(ctx context.Context, request PublishRequest) Response {
	return responseSuccess()
}
