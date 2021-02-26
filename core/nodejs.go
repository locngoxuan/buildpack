package core

import (
	"encoding/json"
	"fmt"
	"github.com/locngoxuan/buildpack/utils"
	"io/ioutil"
	"os"
)

type PackageJson struct {
	Package string
	Name    string
	Version string
}

func ReadPackageJson(file string) (PackageJson, error) {
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		return PackageJson{}, fmt.Errorf("%s file not found", file)
	}

	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		return PackageJson{}, fmt.Errorf("read application config file get error %v", err)
	}

	var packageJson PackageJson
	err = json.Unmarshal(yamlFile, &packageJson)
	if err != nil {
		return PackageJson{}, fmt.Errorf("unmarshal application config file get error %v", err)
	}

	if utils.IsStringEmpty(packageJson.Package) {
		return PackageJson{}, fmt.Errorf("package.json is malformed: missing package property")
	}

	if utils.IsStringEmpty(packageJson.Name) {
		return PackageJson{}, fmt.Errorf("package.json is malformed: missing name property")
	}

	return packageJson, nil
}
