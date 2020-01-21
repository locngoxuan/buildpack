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

type PublisherJfrog struct {
	BuildSnapshot bool
	BuildPack
	PublishJfrogOption
}

type PublishJfrogOption struct {
}

type JfrogUploadParam struct {
	Url        string
	Repository string
	ModulePath string
	FileName   string
	//local
	Source string
}

func (p *PublisherJfrog) SetBuildPack(bp BuildPack) {
	p.BuildPack = bp
}
func (p *PublisherJfrog) LoadConfig(rtOpt BuildPackModuleRuntimeParams, bp BuildPack) error {
	p.BuildPack = bp
	return nil
}
func (p *PublisherJfrog) Pre() error {
	for _, rtModule := range p.RuntimeParams.Modules {
		
	}
	return nil
}
func (p *PublisherJfrog) Publish() error {
	return nil
}
func (p *PublisherJfrog) Clean() error {
	return nil
}

func (p *PublisherJfrog) uploadFile(param JfrogUploadParam) error {
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
