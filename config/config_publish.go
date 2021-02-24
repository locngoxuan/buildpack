package config

type PublishConfig struct {
	Repos []RepoConfig `yaml:"" json:"repos"`
}

type RepoConfig struct {
	Id        string `yaml:"publisher" json:"id"`
	Publisher string `yaml:"publisher" json:"publisher"`
}
