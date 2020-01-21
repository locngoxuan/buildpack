package main

import (
	"strings"
)

type Publisher interface {
	SetBuildPack(bp BuildPack)
	LoadConfig(rtOpt BuildPackModuleRuntimeParams, bp BuildPack) error
	Pre() error
	Publish() error
	Clean() error
}

type EmptyPublisher struct {
}

func (p *EmptyPublisher) SetBuildPack(bp BuildPack) {

}

func (p *EmptyPublisher) LoadConfig(rtOpt BuildPackModuleRuntimeParams, bp BuildPack) error {
	return nil
}
func (p *EmptyPublisher) Pre() error {
	return nil
}
func (p *EmptyPublisher) Publish() error {
	return nil
}
func (p *EmptyPublisher) Clean() error {
	return nil
}

var publishers map[string]Publisher

const (
	publisherJfrogMvn = "jfrog-mvn"
)

func init() {
	publishers = make(map[string]Publisher)
	publishers[publisherJfrogMvn] = &PublisherJfrog{}
}

func publisherOptions() string {
	names := make([]string, 0)
	for name, _ := range publishers {
		names = append(names, name)
	}
	return strings.Join(names, "/")
}

func getPublisher(builderName string) Publisher {
	publisher, ok := publishers[builderName]
	if !ok {
		return nil
	}
	return publisher
}
