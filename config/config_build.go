package config

/**
Example:

builder: mvn/yarn/sql/custom.{name}
image: {docker_image_name}
label: SNAPSHOT
output:
  - target
  - dist
  - libs
 */

type BuildConfig struct {
	Type        string   `yaml:"type,omitempty"`
	DockerImage string   `yaml:"image,omitempty"`
	Label       string   `yaml:"label,omitempty"`
	Output      []string `yaml:"output,omitempty"`
}
