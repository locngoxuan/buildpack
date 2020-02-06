package buildpack

import (
	"fmt"
	"os"
	"strings"
)

const (
	RepoUserPattern     = "REPO_%s_USER"
	RepoPasswordPattern = "REPO_%s_PASS"
	RepoTokenPattern    = "REPO_%s_TOKEN"
	GitToken            = "GIT_TOKEN"
)

func FormatKey(format string, args... string) string{
	return fmt.Sprintf(format, args)
}

func ReadEnvByUpperKey(key string) string{
	return strings.TrimSpace(os.Getenv(strings.ToUpper(key)))
}

func ReadEnv(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}
