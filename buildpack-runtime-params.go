package main

type BuildPackRuntimeParams struct {
	Action            string
	Version           string
	ArtifactoryConfig ArtifactoryConfig
	GitConfig         GitConfig
	DockerConfig      DockerConfig
	Modules           []BuildPackModule
}
