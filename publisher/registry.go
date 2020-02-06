package publisher

var publishTools map[string]PublishTool

func init() {
	publishTools = make(map[string]PublishTool)
	publishTools[artifactoryMvnTool] = &ArtifactoryMVNTool{}
}
