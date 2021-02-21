package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type PackageJson struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func ReadNodeJSPackageJson(file string) (pj PackageJson, err error) {
	_, err = os.Stat(file)
	if os.IsNotExist(err) {
		err = errors.New(file + " file not found")
		return
	}

	jsonFile, err := ioutil.ReadFile(file)
	if err != nil {
		err = errors.New(fmt.Sprintf("read application config file get error %v", err))
		return
	}

	err = json.Unmarshal(jsonFile, &pj)
	if err != nil {
		err = errors.New(fmt.Sprintf("unmarshal application config file get error %v", err))
		return
	}

	if len(strings.TrimSpace(pj.Name)) == 0 {
		err = errors.New("missing name information")
		return
	}

	if len(strings.TrimSpace(pj.Version)) == 0 {
		err = errors.New("missing version information")
		return
	}
	return
}
