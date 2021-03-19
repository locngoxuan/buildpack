package config

type PackConfig struct {
	Type          string `yaml:"type,omitempty" json:"type,omitempty"`
	DockerImage   string `yaml:"image,omitempty" json:"image,omitempty"`
	SkipPullImage bool   `default:"false" yaml:"skip_pull_image,omitemptu" json:"image,omitempty"`
}
