package buildpack

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

const pomFile = "pom.xml"

type POM struct {
	XMLName    xml.Name  `xml:"project"`
	Parent     ParentPOM `xml:"parent"`
	GroupId    string    `xml:"groupId"`
	ArtifactId string    `xml:"artifactId"`
	Classifier string    `xml:"packaging"`
	Version    string    `xml:"version"`
}

type ParentPOM struct {
	GroupId    string `xml:"groupId"`
	ArtifactId string `xml:"artifactId"`
	Version    string `xml:"version"`
}

func ReadPOM(pomFile string) (POM, error) {
	_, err := os.Stat(pomFile)
	if os.IsNotExist(err) {
		err = errors.New(pomFile + " file not found")
		return POM{}, err
	}

	yamlFile, err := ioutil.ReadFile(pomFile)
	if err != nil {
		err = errors.New(fmt.Sprintf("read application config file get error %v", err))
		return POM{}, err
	}

	var pomProject POM
	err = xml.Unmarshal(yamlFile, &pomProject)
	if err != nil {
		err = errors.New(fmt.Sprintf("unmarshal application config file get error %v", err))
		return POM{}, err
	}

	if len(strings.TrimSpace(pomProject.GroupId)) == 0 {
		pomProject.GroupId = pomProject.Parent.GroupId
	}

	if len(strings.TrimSpace(pomProject.Version)) == 0 {
		pomProject.Version = pomProject.Parent.Version
	}
	return pomProject, nil
}
