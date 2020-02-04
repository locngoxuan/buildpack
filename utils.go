package buildpack

import "strings"

func IsEmptyString(str string) bool {
	return len(strings.TrimSpace(str)) == 0
}
