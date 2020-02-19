package builder

var buildTools map[string]BuildTool

func init() {
	buildTools = make(map[string]BuildTool)
	buildTools[mvnBuildTool] = &MVNBuildTool{}
	buildTools[mvnAppBuildTool] = &MVNAppBuildTool{}
	buildTools[sqlBundleBuildTool] = &SQLBundleBuildTool{}
	buildTools[mvnAutoBuildTool] = &MVNAutoBuildTool{}
	buildTools[noBuildTool] = &NoBuild{}
}

func Listed() []string {
	list := make([]string, 0)
	for key, _ := range buildTools {
		list = append(list, key)
	}
	return list
}
