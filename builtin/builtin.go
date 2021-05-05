package builtin

import "github.com/locngoxuan/buildpack/instrument"

func InitBuiltInFunction(){
	instrument.RegisterBuildDockerImage(MvnBuilderName, defaultMvnDockerImage)
	instrument.RegisterBuildFunction(MvnBuilderName, mvnBuild)
	instrument.RegisterBuildDockerImage(NpmBuilderName, defaultNodeLtsDockerImage)
	instrument.RegisterBuildFunction(NpmBuilderName, npmBuild)
	instrument.RegisterBuildDockerImage(YarnBuilderName, defaultNodeLtsDockerImage)
	instrument.RegisterBuildFunction(YarnBuilderName, yarnBuild)

	instrument.RegisterPackDockerImage(NpmPackerName, defaultNodeLtsDockerImage)
	instrument.RegisterPackFunction(NpmPackerName, npmPack)
	instrument.RegisterPackDockerImage(YarnPackerName, defaultNodeLtsDockerImage)
	instrument.RegisterPackFunction(YarnPackerName, yarnPack)

	instrument.RegisterPublishFunction(ArtifactoryMvnPublisherName, publishMvnJarToArtifactory)
	instrument.RegisterPublishFunction(ArtifactoryYarnPublisherName, publishYarnJarToArtifactory)
	instrument.RegisterPublishFunction(ArtifactoryNpmPublisherName, publishYarnJarToArtifactory)
}
