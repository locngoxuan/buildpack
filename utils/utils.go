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
	origin := Trim(str)
	if strings.HasPrefix(origin, "$") {
		result := os.ExpandEnv(origin)
		if !IsStringEmpty(result){
			return result
		}
	}
	return origin
}
