package common

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
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

func SumContentMD5(file string) (string, error) {
	hasher := md5.New()
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = f.Close()
	}()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}
