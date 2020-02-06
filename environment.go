package buildpack

import (
	"os"
	"strings"
)

const (
	RepoUserPattern     = "REPO_%s_USER"
	RepoPasswordPattern = "REPO_%s_PASS"
	RepoTokenPattern    = "REPO_%s_TOKEN"
	GitToken            = "GIT_TOKEN"
)

func ReadEnv(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}
