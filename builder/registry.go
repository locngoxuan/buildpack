package builder

var buildTools map[string]BuildTool

func init() {
	buildTools = make(map[string]BuildTool)
	buildTools[mvnBuildTool] = &MVNBuildTool{}
}
