package common

import (
	"errors"
	"os"
	"strings"
)

const AlpineImage = "alpine:3.12.0"

func IsEmptyString(s string) bool {
	return strings.TrimSpace(s) == ""
}

func CreateDir(dir string, skipContainer bool, perm os.FileMode) error {
	if !skipContainer {
		return errors.New("not implemented yet")
	} else {
		return os.MkdirAll(dir, perm)
	}
}

func DeleteDir(dir string, skipContainer bool) error {
	if !skipContainer {
		return errors.New("not implemented yet")
	} else {
		return os.RemoveAll(dir)
	}
}
