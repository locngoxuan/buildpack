package main

import (
	"os"
	"strings"
)

func isStringEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

func isNotExists(s string) bool{
	_, err := os.Stat(s)
	return os.IsNotExist(err)
}