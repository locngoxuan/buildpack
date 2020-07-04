package publisher

type PublishContext struct {
	Name      string
	Path      string
	WorkDir   string
	OutputDir string
	RepoName  string
	Version   string
	IsStable  bool
}
