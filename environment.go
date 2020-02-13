package buildpack

import (
	"fmt"
	"os"
	"strings"
)

const (
	RepoUserPattern     = "REPO_%s_USER"
	RepoPasswordPattern = "REPO_%s_PASS"
	GitToken            = "GIT_TOKEN"
)

func FormatKey(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

func ReadEnvByUpperKey(key string) string {
	return strings.TrimSpace(os.Getenv(strings.ToUpper(key)))
}

func ReadEnv(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}
