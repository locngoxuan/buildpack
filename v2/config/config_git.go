package config

type GitConfig struct {
	Branch        string `yaml:"branch,omitempty" json:"branch,omitempty"`
	RemoteAddress string `yaml:"remote,omitempty" json:"remote,omitempty"`
	GitCredential `yaml:"credential,omitempty" json:"credential,omitempty"`
}

const (
	CredentialToken   = "token"
	CredentialAccount = "account"
)

type GitCredential struct {
	Type        string `yaml:"type,omitempty" json:"type,omitempty"`
	AccessToken string `yaml:"access_token,omitempty" json:"access_token,omitempty"`
	Username    string `yaml:"username,omitempty" json:"username,omitempty"`
	Password    string `yaml:"password,omitempty" json:"password,omitempty"`
}
