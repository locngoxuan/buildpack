package buildpack

type GitManager struct {
	DisplayName string
	Username    string
	Email       string
	Password    string
}

func CreateGitManager() GitManager{
	return GitManager{}
}