package core

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type POM struct {
	XMLName    xml.Name  `xml:"project"`
	Parent     ParentPOM `xml:"parent"`
	GroupId    string    `xml:"groupId"`
	ArtifactId string    `xml:"artifactId"`
	Classifier string    `xml:"packaging"`
	Version    string    `xml:"version"`
	Build      BuildTag  `xml:"build"`
}

type ParentPOM struct {
	GroupId    string `xml:"groupId"`
	ArtifactId string `xml:"artifactId"`
	Version    string `xml:"version"`
}

type BuildTag struct {
	FinalName string `xml:"finalName"`
}

func ReadPOM(pomFile string) (POM, error) {
	_, err := os.Stat(pomFile)
	if os.IsNotExist(err) {
		return POM{}, fmt.Errorf("%s file not found",pomFile)
	}

	yamlFile, err := ioutil.ReadFile(pomFile)
	if err != nil {
		return POM{}, fmt.Errorf("read application config file get error %v", err)
	}

	var pomProject POM
	err = xml.Unmarshal(yamlFile, &pomProject)
	if err != nil {
		return POM{}, fmt.Errorf("unmarshal application config file get error %v", err)
	}

	if len(strings.TrimSpace(pomProject.GroupId)) == 0 {
		pomProject.GroupId = pomProject.Parent.GroupId
	}

	if len(strings.TrimSpace(pomProject.Version)) == 0 {
		pomProject.Version = pomProject.Parent.Version
	}
	return pomProject, nil
}
