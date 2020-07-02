package common

import (
	"errors"
	"os"
	"strings"
)

func IsEmptyString(s string) bool {
	return strings.TrimSpace(s) == ""
}

func CreateDir(dir string, skipContainer bool) error {
	if !skipContainer {
		return errors.New("not implemented yet")
	} else {
		return os.MkdirAll(dir, 0755)
	}
}

func DeleteDir(dir string, skipContainer bool) error {
	if !skipContainer {
		return errors.New("not implemented yet")
	} else {
		return os.RemoveAll(dir)
	}
}
