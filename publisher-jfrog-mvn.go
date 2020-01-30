package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
)

const (
	publishDir = "publishes"
)

type PublisherJfrogMvn struct {
	BuildSnapshot bool
	BuildPack
	PublishJfrogOption
}

type PublishJfrogOption struct {
	GroupId    string `yaml:"group,omitempty"`
	ArtifactId string `yaml:"artifact,omitempty"`
	Classifier string `yaml:"classifier,omitempty"`
	Version    string `yaml:"version,omitempty"`
	Label      string `yaml:"label,omitempty"`
}

type JfrogUploadParam struct {
	Url        string
	Repository string
	ModulePath string
	FileName   string
	//local
	Source string
}

func (p *PublisherJfrogMvn) WriteConfig(name, path string, opt BuildPackModuleConfig) error {
	return nil
}

func (p *PublishJfrogOption) Verify() error {
	return nil
}

func (p *PublisherJfrogMvn) SetBuildPack(bp BuildPack) {
	p.BuildPack = bp
}
func (p *PublisherJfrogMvn) LoadConfig(rtOpt BuildPackModuleRuntimeParams, bp BuildPack) error {
	p.BuildPack = bp
	return nil
}
func (p *PublisherJfrogMvn) Pre() error {
	for _, rtModule := range p.RuntimeParams.Modules {
		pomSrc := p.BuildPack.buildPathOnRoot(rtModule.Path, "target", pomFlattened)
		pomPublished := p.BuildPack.buildPathOnRoot(publishDir, fmt.Sprintf(".pom"))
		err := copyFile(pomSrc, pomPublished)
		if err != nil {
			return err
		}
	}
	return nil
}
func (p *PublisherJfrogMvn) Publish() error {
	return nil
}
func (p *PublisherJfrogMvn) Clean() error {
	return nil
}

func (p *PublisherJfrogMvn) uploadFile(param JfrogUploadParam) error {
	destination := fmt.Sprintf("%s/%s/%s/%s", param.Url, param.Repository, param.ModulePath, param.FileName)
	buildInfo(p.BuildPack, fmt.Sprintf("PUT %s to %s", param.Source, destination))
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
	req.SetBasicAuth(p.BuildPack.RuntimeParams.Username, p.BuildPack.RuntimeParams.Password)

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
