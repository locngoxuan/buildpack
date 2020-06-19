package buildpack

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

const packageJson = "package.json"

type PackageJson struct {
	Package string `json:"package"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

func ReadPackageJson(file string) (PackageJson, error) {
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		err = errors.New(file + " file not found")
		return PackageJson{}, err
	}

	jsonFile, err := ioutil.ReadFile(packageJson)
	if err != nil {
		err = errors.New(fmt.Sprintf("read application config file get error %v", err))
		return PackageJson{}, err
	}

	var pomProject PackageJson
	err = json.Unmarshal(jsonFile, &pomProject)
	if err != nil {
		err = errors.New(fmt.Sprintf("unmarshal application config file get error %v", err))
		return PackageJson{}, err
	}

	if len(strings.TrimSpace(pomProject.Package)) == 0 {
		return PackageJson{}, errors.New("missing package information")
	}

	if len(strings.TrimSpace(pomProject.Name)) == 0 {
		return PackageJson{}, errors.New("missing name information")
	}

	if len(strings.TrimSpace(pomProject.Version)) == 0 {
		return PackageJson{}, errors.New("missing version information")
	}
	return pomProject, nil
}
