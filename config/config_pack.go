package config

type PackConfig struct {
	Type        string `yaml:"type,omitempty" json:"type,omitempty"`
	DockerImage string `yaml:"image,omitempty" json:"image,omitempty"`
}
