package config

//publish config in each module
type PublishConfig struct {
	Type    string   `yaml:"type,omitempty" json:"type,omitempty"`
	RepoIds []string `yaml:"repo_ids,omitempty" json:"repo_ids,omitempty"`
}

//repository
type RepositoryConfig struct {
	Repos []Repository `yaml:"repositories,omitempty" json:"repositories,omitempty"`
}

type Repository struct {
	Id         string  `yaml:"id,omitempty" json:"id,omitempty"`
	DevChannel Channel `yaml:"channel_dev,omitempty" json:"channel_dev,omitempty"`
	RelChannel Channel `yaml:"channel_rel,omitempty" json:"channel_rel,omitempty"`
}

func (r Repository) GetChannel(release bool) Channel {
	if release {
		return r.RelChannel
	}
	return r.DevChannel
}

type Channel struct {
	Address  string `yaml:"address,omitempty" json:"address,omitempty"`
	Username string `yaml:"username,omitempty" json:"username,omitempty"`
	Password string `yaml:"password,omitempty" json:"password,omitempty"`
}
