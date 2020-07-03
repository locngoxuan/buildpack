package buildpack

type GitManager struct {
	DisplayName string
	Username    string
	Email       string
	Password    string
}

func CreateGitManager(c BuildConfig) GitManager {
	return GitManager{}
}
