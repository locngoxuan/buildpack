package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

const (
	publishDir = "publishes"
)

type PublisherJfrogMvn struct {
	PublishJfrogOption
}

type PublishJfrogOption struct {
	POM
}

type JfrogUploadParam struct {
	Url        string
	Repository string
	ModulePath string
	FileName   string
	//local
	Source string
}

type POM struct {
	XMLName    xml.Name `xml:"project"`
	GroupId    string   `xml:"groupId"`
	ArtifactId string   `xml:"artifactId"`
	Classifier string   `xml:"packaging"`
}

func (p *PublisherJfrogMvn) WriteConfig(bp BuildPack, opt BuildPackModuleConfig) error {
	return nil
}

func (p *PublisherJfrogMvn) CreateContext(bp BuildPack, rtOpt BuildPackModuleRuntimeParams) (PublishContext, error) {
	ctx := PublishContext{}
	pwd, err := filepath.Abs(bp.getModuleWorkingDir(rtOpt.Path))
	if err != nil {
		return ctx, err
	}
	configFile := filepath.Join(pwd, pomFile)
	_, err = os.Stat(configFile)
	if os.IsNotExist(err) {
		err = errors.New("configuration file not found")
		return ctx, err
	}

	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		err = errors.New(fmt.Sprintf("read application config file get error %v", err))
		return ctx, err
	}

	var pomProject POM
	err = xml.Unmarshal(yamlFile, &pomProject)
	if err != nil {
		err = errors.New(fmt.Sprintf("unmarshal application config file get error %v", err))
		return ctx, err
	}
	ctx.Name = rtOpt.Name
	ctx.Path = rtOpt.Path
	p.POM = pomProject
	return ctx, nil
}

func (p *PublishJfrogOption) Verify(ctx PublishContext) error {
	return nil
}

func (p *PublisherJfrogMvn) Pre(ctx PublishContext) error {
	rtModule := ctx.BuildPackModuleRuntimeParams
	pomSrc := ctx.buildPathOnRoot(rtModule.Path, pomFlattened)

	version := ctx.RuntimeParams.VersionRuntimeParams.version(labelSnapshot, 0)
	pomName := fmt.Sprintf("%s-%s.pom", p.ArtifactId, version)
	pomPublished := ctx.buildPathOnRoot(publishDir, pomName)
	err := copyFile(pomSrc, pomPublished)
	if err != nil {
		return err
	}
	return nil
}

func (p *PublisherJfrogMvn) Publish(ctx PublishContext) error {
	return nil
}

func (p *PublisherJfrogMvn) Clean(ctx PublishContext) error {
	return nil
}

func uploadFile(bp BuildPack, param JfrogUploadParam) error {
	destination := fmt.Sprintf("%s/%s/%s/%s", param.Url, param.Repository, param.ModulePath, param.FileName)
	buildInfo(bp, fmt.Sprintf("PUT %s to %s", param.Source, destination))
	data, err := os.Open(param.Source)
	if err != nil {
		return err
	}
	defer func() {
		_ = data.Close()
	}()
	req, err := http.NewRequest("PUT", destination, data)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")
	req.SetBasicAuth(bp.RuntimeParams.Username, bp.RuntimeParams.Password)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = res.Body.Close()
	}()
	if res.StatusCode != http.StatusCreated {
		return errors.New(res.Status)
	}
	return nil
}
