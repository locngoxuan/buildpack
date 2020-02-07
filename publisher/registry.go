package publisher

var publishTools map[string]PublishTool

func init() {
	publishTools = make(map[string]PublishTool)
	publishTools[artifactoryMvnTool] = &ArtifactoryMVNTool{}
}

func Listed() []string {
	list := make([]string, 0)
	for key, _ := range publishTools {
		list = append(list, key)
	}
	return list
}
