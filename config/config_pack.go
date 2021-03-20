package config

type PackConfig struct {
	Type             string `yaml:"type,omitempty" json:"type,omitempty"`
	DockerImage      string `yaml:"image,omitempty" json:"image,omitempty"`
	SkipPrepareImage bool   `default:"false" yaml:"skip_prepare,omitempty" json:"image,omitempty"`
}
