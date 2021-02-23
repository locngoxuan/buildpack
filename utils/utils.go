package utils

import (
	"os"
	"strings"
)

func IsStringEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

func Trim(s string) string {
	return strings.TrimSpace(s)
}

func IsNotExists(s string) bool {
	_, err := os.Stat(s)
	return os.IsNotExist(err)
}

func ReadEnvVariableIfHas(str string) string {
	result := strings.TrimSpace(str)
	if strings.HasPrefix(result, "$") {
		result = os.ExpandEnv(result)
	}
	return result
}
